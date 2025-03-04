package mongodb

import (
	"context"
	"fmt"
	"math"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoBlockStore) SetBlockInfo(ctx context.Context, blockInfo *domain.BlockInfo) error {
	filter := bson.M{partitionIDKey: blockInfo.PartitionID, blockNumberKey: blockInfo.BlockNumber}
	update := bson.M{"$set": blockInfo}

	_, err := s.db.Collection(blocksCollectionName).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to upsert blockInfo: %w", err)
	}
	return nil
}

func (s *MongoBlockStore) GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error) {
	filter := bson.M{blockNumberKey: blockNumber}
	if len(partitionIDs) > 0 {
		filter[partitionIDKey] = bson.M{"$in": partitionIDs}
	}

	cursor, err := s.db.Collection(blocksCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query blocks: %w", err)
	}
	defer cursor.Close(ctx)

	blockMap := make(map[types.PartitionID]*domain.BlockInfo)
	for cursor.Next(ctx) {
		var block domain.BlockInfo
		if err = cursor.Decode(&block); err != nil {
			return nil, fmt.Errorf("failed to decode block: %w", err)
		}
		blockMap[block.PartitionID] = &block
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor encountered an error: %w", err)
	}

	return blockMap, nil
}

func (s *MongoBlockStore) GetLastBlocks(
	ctx context.Context,
	partitionIDs []types.PartitionID,
	count int,
	includeEmpty bool,
) (map[types.PartitionID][]*domain.BlockInfo, error) {
	blockMap := make(map[types.PartitionID][]*domain.BlockInfo)

	if len(partitionIDs) == 0 {
		latestBlockNumbers, err := s.GetBlockNumbers(ctx, nil)
		if err != nil {
			return nil, err
		}
		for partitionID, _ := range latestBlockNumbers {
			partitionIDs = append(partitionIDs, partitionID)
		}
	}

	for _, partitionID := range partitionIDs {
		blocks, _, err := s.GetBlocksInRange(ctx, partitionID, math.MaxInt64, count, includeEmpty)
		if err != nil {
			return nil, fmt.Errorf("failed to get blocks for partition %d: %w", partitionID, err)
		}
		blockMap[partitionID] = blocks
	}

	return blockMap, nil
}

func (s *MongoBlockStore) GetBlocksInRange(
	ctx context.Context,
	partitionID types.PartitionID,
	latestBlock uint64,
	count int,
	includeEmpty bool,
) ([]*domain.BlockInfo, uint64, error) {
	filter := bson.M{
		partitionIDKey: partitionID,
		blockNumberKey: bson.M{"$lte": latestBlock},
	}

	if !includeEmpty {
		filter[txCountKey] = bson.M{
			"$gt": 0,
		}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: blockNumberKey, Value: -1}}).
		SetLimit(int64(count))

	cursor, err := s.db.Collection(blocksCollectionName).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query blocks: %w", err)
	}
	defer cursor.Close(ctx)

	var blocks []*domain.BlockInfo
	for cursor.Next(ctx) {
		var block domain.BlockInfo
		if err = cursor.Decode(&block); err != nil {
			return nil, 0, fmt.Errorf("failed to decode block: %w", err)
		}
		blocks = append(blocks, &block)
	}

	if len(blocks) == 0 {
		return blocks, 0, nil
	}

	prevBlockNumber := uint64(0)
	if len(blocks) == count {
		prevBlockNumber = blocks[len(blocks)-1].BlockNumber - 1
	}

	return blocks, prevBlockNumber, nil
}
