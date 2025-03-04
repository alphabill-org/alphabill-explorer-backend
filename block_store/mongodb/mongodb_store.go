package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/alphabill-org/alphabill-go-base/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName           = "blockExplorerDB"
	blocksCollectionName   = "blocks"
	txCollectionName       = "transactions"
	metadataCollectionName = "metadata"

	partitionIDKey       = "partitionid"
	blockNumberKey       = "blocknumber"
	txRecordHashKey      = "txrecordhash"
	txOrderHashKey       = "txorderhash"
	txHashesKey          = "txhashes"
	txCountKey           = "txcount"
	targetUnitsKey       = "transaction.servermetadata.targetunits"
	latestBlockNumberKey = "latestblocknumber"

	connectTimeout       = time.Minute
	connectionRetries    = 5
	connectionRetryDelay = 5 * time.Second
)

type MongoBlockStore struct {
	db *mongo.Database
}

func NewMongoBlockStore(ctx context.Context, uri string) (*MongoBlockStore, error) {
	for i := 0; ; i++ {
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri), options.Client().SetConnectTimeout(connectTimeout))
		if err != nil {
			if i == connectionRetries {
				return nil, fmt.Errorf("failed to connect to mongo: %w", err)
			}
			fmt.Printf("Failed to connect to mongo, retrying after %v... err = %s\n", connectionRetryDelay, err)
			time.Sleep(connectionRetryDelay)
			continue
		}
		store := &MongoBlockStore{db: client.Database(databaseName)}
		if err = store.initialize(ctx); err != nil {
			return nil, err
		}
		return store, nil
	}
}

// ensureCollectionExists creates the collection if it doesn't exist
func ensureCollectionExists(ctx context.Context, db *mongo.Database, collectionName string) error {
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": collectionName})
	if err != nil {
		return fmt.Errorf("failed to list collections: %v", err)
	}

	if len(collections) == 0 {
		err := db.CreateCollection(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("failed to create collection '%s': %v", collectionName, err)
		}
		fmt.Printf("Created collection: %s\n", collectionName)
	} else {
		fmt.Printf("Collection '%s' already exists\n", collectionName)
	}
	return nil
}

func createMetadataCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection(metadataCollectionName)

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: partitionIDKey, Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index on metadata collection: %v", err)
	}
	return nil
}

func createIndexes(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection(blocksCollectionName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: blockNumberKey, Value: 1}, {Key: partitionIDKey, Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: partitionIDKey, Value: 1}, {Key: blockNumberKey, Value: -1}}, // for GetLatestBlock query
		},
		{
			Keys: bson.D{{Key: partitionIDKey, Value: 1}, {Key: txCountKey, Value: 1}, {Key: blockNumberKey, Value: -1}},
		},
	})
	if err != nil {
		return err
	}

	_, err = db.Collection(txCollectionName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: txRecordHashKey, Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: txOrderHashKey, Value: 1}},
		},
		{Keys: bson.D{
			{Key: partitionIDKey, Value: 1},
			{Key: "_id", Value: -1},
		}},
		{Keys: bson.D{{Key: targetUnitsKey, Value: 1}}},
		{Keys: bson.D{{Key: blockNumberKey, Value: 1}, {Key: partitionIDKey, Value: 1}}},
	})
	return err
}

func (s *MongoBlockStore) GetBlockNumber(ctx context.Context, partitionID types.PartitionID) (uint64, error) {
	blockNumberMap, err := s.GetBlockNumbers(ctx, []types.PartitionID{partitionID})
	if err != nil {
		return 0, err
	}
	return blockNumberMap[partitionID], nil
}

func (s *MongoBlockStore) GetBlockNumbers(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]uint64, error) {
	var filter bson.M

	// If no partitions specified, retrieve all partitions
	if len(partitionIDs) == 0 {
		filter = bson.M{}
	} else {
		filter = bson.M{partitionIDKey: bson.M{"$in": partitionIDs}}
	}

	cursor, err := s.db.Collection(metadataCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest block numbers: %w", err)
	}
	defer cursor.Close(ctx)

	blockNumbers := make(map[types.PartitionID]uint64)
	for cursor.Next(ctx) {
		var result struct {
			PartitionID       types.PartitionID `bson:"partitionid"`
			LatestBlockNumber uint64            `bson:"latestblocknumber"`
		}

		if err = cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode block number: %w", err)
		}

		blockNumbers[result.PartitionID] = result.LatestBlockNumber
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor encountered an error: %w", err)
	}

	// If partitionIDs is not nil, ensure all requested partitions are accounted for
	if partitionIDs != nil {
		for _, partitionID := range partitionIDs {
			if _, found := blockNumbers[partitionID]; !found {
				if err = s.SetBlockNumber(ctx, partitionID, 0); err != nil {
					return nil, fmt.Errorf("failed to set default block number for partition %s: %w", partitionID, err)
				}
				blockNumbers[partitionID] = 0
			}
		}
	}

	return blockNumbers, nil
}

func (s *MongoBlockStore) SetBlockNumber(ctx context.Context, partitionID types.PartitionID, blockNumber uint64) error {
	filter := bson.M{partitionIDKey: partitionID}
	update := bson.M{
		"$set": bson.M{
			partitionIDKey:       partitionID,
			latestBlockNumberKey: blockNumber,
		},
	}

	_, err := s.db.Collection(metadataCollectionName).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to set block number: %w", err)
	}
	return nil
}

func (s *MongoBlockStore) ResetCollections(ctx context.Context) error {
	if err := s.db.Collection(blocksCollectionName).Drop(ctx); err != nil {
		return err
	}
	if err := s.db.Collection(txCollectionName).Drop(ctx); err != nil {
		return err
	}
	if err := s.db.Collection(metadataCollectionName).Drop(ctx); err != nil {
		return err
	}
	return s.initialize(ctx)
}

func (s *MongoBlockStore) initialize(ctx context.Context) error {
	if err := ensureCollectionExists(ctx, s.db, blocksCollectionName); err != nil {
		return err
	}
	if err := ensureCollectionExists(ctx, s.db, txCollectionName); err != nil {
		return err
	}
	if err := createMetadataCollection(ctx, s.db); err != nil {
		return err
	}
	if err := createIndexes(ctx, s.db); err != nil {
		return err
	}
	if err := s.MigrateTxCount(ctx); err != nil {
		return fmt.Errorf("failed to run migration: %w", err)
	}

	return nil
}

// todo remove
func (s *MongoBlockStore) MigrateTxCount(ctx context.Context) error {
	fmt.Println("Starting migration: Adding txCount to existing blocks...")

	filter := bson.M{txCountKey: bson.M{"$exists": false}}

	update := bson.A{ // Use an aggregation pipeline for $size
		bson.M{"$set": bson.M{txCountKey: bson.M{"$size": fmt.Sprintf("$%s", txHashesKey)}}},
	}

	result, err := s.db.Collection(blocksCollectionName).UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to migrate txCount field: %w", err)
	}

	fmt.Printf("Migration complete: Updated %d blocks with txCount\n", result.ModifiedCount)
	return nil
}
