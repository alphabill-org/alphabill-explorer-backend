package mongodb

import (
	"context"
	"fmt"
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
	targetUnitsKey       = "transaction.servermetadata.targetunits"
	latestBlockNumberKey = "latestblocknumber"
)

type MongoBillStore struct {
	db *mongo.Database
}

func NewMongoBillStore(ctx context.Context, uri string) (*MongoBillStore, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	store := &MongoBillStore{db: client.Database(databaseName)}
	if err = store.createCollections(ctx); err != nil {
		return nil, err
	}
	return store, nil
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
		{Keys: bson.D{{Key: partitionIDKey, Value: 1}, {Key: blockNumberKey, Value: -1}}}, // for GetLatestBlock query
	})
	if err != nil {
		return err
	}

	_, err = db.Collection(txCollectionName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: txRecordHashKey, Value: 1}},
			Options: options.Index().SetUnique(true),
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

func (s *MongoBillStore) GetBlockNumber(ctx context.Context, partitionID types.PartitionID) (uint64, error) {
	filter := bson.M{partitionIDKey: partitionID}

	var result struct {
		PartitionID       types.PartitionID
		LatestBlockNumber uint64
	}

	err := s.db.Collection(metadataCollectionName).FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return 0, s.SetBlockNumber(ctx, partitionID, 0)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %w", err)
	}

	return result.LatestBlockNumber, nil
}

func (s *MongoBillStore) SetBlockNumber(ctx context.Context, partitionID types.PartitionID, blockNumber uint64) error {
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

func (s *MongoBillStore) ResetCollections(ctx context.Context) error {
	if err := s.db.Collection(blocksCollectionName).Drop(ctx); err != nil {
		return err
	}
	if err := s.db.Collection(txCollectionName).Drop(ctx); err != nil {
		return err
	}
	if err := s.db.Collection(metadataCollectionName).Drop(ctx); err != nil {
		return err
	}
	return s.createCollections(ctx)
}

func (s *MongoBillStore) createCollections(ctx context.Context) error {
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
	return nil
}
