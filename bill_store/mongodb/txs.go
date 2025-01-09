package mongodb

import (
	"context"
	"fmt"
	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
	"go.mongodb.org/mongo-driver/bson"
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

// GetTxsInRange returns a list of transactions starting from the given sequence number. startSequenceNumber=0 means it's not set and cursor.last() is used.
func (s *MongoBillStore) GetTxsInRange(ctx context.Context, partitionID types.PartitionID, startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error) {
	// todo
	return nil, 0, nil
}
