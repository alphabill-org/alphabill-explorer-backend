package mongodb

import (
	"context"
	"fmt"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoBlockStore) SetTxInfo(ctx context.Context, txInfo *domain.TxInfo) error {
	filter := bson.M{txRecordHashKey: txInfo.TxRecordHash}
	update := bson.M{"$set": txInfo}

	_, err := s.db.Collection(txCollectionName).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to upsert transaction: %w", err)
	}
	return nil
}

func (s *MongoBlockStore) GetTxByHash(ctx context.Context, txHash domain.TxHash) (*domain.TxInfo, error) {
	filter := bson.M{
		"$or": []bson.M{
			{txRecordHashKey: txHash},
			{txOrderHashKey: txHash},
		},
	}

	var tx domain.TxInfo
	err := s.db.Collection(txCollectionName).FindOne(ctx, filter).Decode(&tx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query transaction by hash: %w", err)
	}

	return &tx, nil
}

func (s *MongoBlockStore) GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*domain.TxInfo, error) {
	blockMap, err := s.GetBlock(ctx, blockNumber, []types.PartitionID{partitionID})
	if err != nil {
		return nil, err
	}
	if blockMap == nil || blockMap[partitionID] == nil {
		return nil, fmt.Errorf("could not find block with number %d in partition %d", blockNumber, partitionID)
	}
	return s.getTxsByHashes(ctx, blockMap[partitionID].TxHashes)
}

func (s *MongoBlockStore) getTxsByHashes(ctx context.Context, hashes []domain.TxHash) ([]*domain.TxInfo, error) {
	filter := bson.M{txRecordHashKey: bson.M{"$in": hashes}}

	cursor, err := s.db.Collection(txCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions by hashes: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*domain.TxInfo
	for cursor.Next(ctx) {
		var tx domain.TxInfo
		if err = cursor.Decode(&tx); err != nil {
			return nil, fmt.Errorf("failed to decode transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor encountered an error: %w", err)
	}

	return transactions, nil
}

func (s *MongoBlockStore) GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*domain.TxInfo, error) {
	filter := bson.M{targetUnitsKey: unitID}

	cursor, err := s.db.Collection(txCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions by unitID: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*domain.TxInfo
	for cursor.Next(ctx) {
		var tx domain.TxInfo
		if err := cursor.Decode(&tx); err != nil {
			return nil, fmt.Errorf("failed to decode transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor encountered an error: %w", err)
	}

	return transactions, nil
}

// GetTxsPage retrieves a paginated list of transactions for a given partition, starting from the specified latestID.
// Returns the transactions, the latest ID for the previous page, and any error encountered.
func (s *MongoBlockStore) GetTxsPage(
	ctx context.Context,
	partitionID types.PartitionID,
	latestID string,
	limit int,
) (transactions []*domain.TxInfo, previousID string, err error) {
	filter := bson.M{partitionIDKey: partitionID}
	if latestID != "" {
		objectID, err := primitive.ObjectIDFromHex(latestID)
		if err != nil {
			return nil, "", fmt.Errorf("invalid startID: %w", err)
		}
		filter["_id"] = bson.M{"$lte": objectID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(int64(limit + 1)) // Fetch one extra to identify the previous ID

	cursor, err := s.db.Collection(txCollectionName).Find(ctx, filter, opts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query transactions: %w", err)
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var tx domain.TxInfo
		if err = cursor.Decode(&tx); err != nil {
			return nil, "", fmt.Errorf("failed to decode transaction: %w", err)
		}

		if count == limit {
			previousID = tx.ID.Hex()
			break
		}

		transactions = append(transactions, &tx)
		count++
	}

	if err = cursor.Err(); err != nil {
		return nil, "", fmt.Errorf("cursor encountered an error: %w", err)
	}

	return transactions, previousID, nil
}

func (s *MongoBlockStore) FindTxs(ctx context.Context, searchKey []byte, partitionIDs []types.PartitionID) ([]*domain.TxInfo, error) {
	filter := bson.M{
		"$or": []bson.M{
			{txRecordHashKey: searchKey},
			{txOrderHashKey: searchKey},
			{targetUnitsKey: searchKey},
		},
	}
	if len(partitionIDs) > 0 {
		filter[partitionIDKey] = bson.M{"$in": partitionIDs}
	}

	cursor, err := s.db.Collection(txCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*domain.TxInfo
	for cursor.Next(ctx) {
		var tx domain.TxInfo
		if err := cursor.Decode(&tx); err != nil {
			return nil, fmt.Errorf("failed to decode transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor encountered an error: %w", err)
	}

	if len(transactions) == 0 {
		return nil, domain.ErrNotFound
	}

	return transactions, nil
}
