package mongodb

import (
	"context"
	"fmt"
	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoBillStore) SetTxInfo(ctx context.Context, txInfo *api.TxInfo) error {
	filter := bson.M{txRecordHashKey: txInfo.TxRecordHash}
	update := bson.M{"$set": txInfo}

	_, err := s.db.Collection(txCollectionName).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to upsert transaction: %w", err)
	}
	return nil
}

func (s *MongoBillStore) GetTxInfo(ctx context.Context, txHash api.TxHash) (*api.TxInfo, error) {
	filter := bson.M{txRecordHashKey: txHash}

	var tx api.TxInfo
	err := s.db.Collection(txCollectionName).FindOne(ctx, filter).Decode(&tx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("transaction not found for hash: %x", txHash)
		}
		return nil, fmt.Errorf("failed to query transaction by hash: %w", err)
	}

	return &tx, nil
}

func (s *MongoBillStore) GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*api.TxInfo, error) {
	blockMap, err := s.GetBlock(ctx, blockNumber, []types.PartitionID{partitionID})
	if err != nil {
		return nil, err
	}
	if blockMap == nil || blockMap[partitionID] == nil {
		return nil, fmt.Errorf("could not find block with number %d in partition %d", blockNumber, partitionID)
	}
	return s.getTxsByHashes(ctx, blockMap[partitionID].TxHashes)
}

func (s *MongoBillStore) getTxsByHashes(ctx context.Context, hashes []api.TxHash) ([]*api.TxInfo, error) {
	filter := bson.M{txRecordHashKey: bson.M{"$in": hashes}}

	cursor, err := s.db.Collection(txCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions by hashes: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*api.TxInfo
	for cursor.Next(ctx) {
		var tx api.TxInfo
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

func (s *MongoBillStore) GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*api.TxInfo, error) {
	filter := bson.M{targetUnitsKey: unitID}

	cursor, err := s.db.Collection(txCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions by unitID: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*api.TxInfo
	for cursor.Next(ctx) {
		var tx api.TxInfo
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

func (s *MongoBillStore) GetTxsPage(
	ctx context.Context,
	partitionID types.PartitionID,
	startID string,
	limit int,
) (transactions []*api.TxInfo, previousID string, err error) {
	filter := bson.M{partitionIDKey: partitionID}
	if startID != "" {
		objectID, err := primitive.ObjectIDFromHex(startID)
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
		var tx api.TxInfo
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
