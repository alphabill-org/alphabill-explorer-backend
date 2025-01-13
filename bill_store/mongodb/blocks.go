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

func (s *MongoBillStore) SetBlockInfo(ctx context.Context, blockInfo *api.BlockInfo) error {
	filter := bson.M{partitionIDKey: blockInfo.PartitionID, blockNumberKey: blockInfo.BlockNumber}
	update := bson.M{"$set": blockInfo}

	_, err := s.db.Collection(blocksCollectionName).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to upsert blockInfo: %w", err)
	}
	return nil
}

func (s *MongoBillStore) GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error) {
	filter := bson.M{blockNumberKey: blockNumber}
	if partitionIDs != nil && len(partitionIDs) > 0 {
		filter[partitionIDKey] = bson.M{"$in": partitionIDs}
	}

	cursor, err := s.db.Collection(blocksCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query blocks: %w", err)
	}
	defer cursor.Close(ctx)

	blockMap := make(map[types.PartitionID]*api.BlockInfo)
	for cursor.Next(ctx) {
		var block api.BlockInfo
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

func (s *MongoBillStore) GetLastBlock(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{Key: partitionIDKey, Value: bson.M{"$in": partitionIDs}},
		}}},
		{{"$sort", bson.D{
			{Key: partitionIDKey, Value: 1},
			{Key: blockNumberKey, Value: -1},
		}}},
		{{"$group", bson.D{
			{Key: "_id", Value: "$" + partitionIDKey},
			{Key: "latestBlockInfo", Value: bson.M{"$first": "$$ROOT"}},
		}}},
	}

	cursor, err := s.db.Collection(blocksCollectionName).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute aggregation: %w", err)
	}
	defer cursor.Close(ctx)

	blockMap := make(map[types.PartitionID]*api.BlockInfo)
	for cursor.Next(ctx) {
		var result struct {
			PartitionID     types.PartitionID `bson:"_id"`
			LatestBlockInfo api.BlockInfo     `bson:"latestBlockInfo"`
		}

		if err = cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode aggregation result: %w", err)
		}

		blockMap[result.PartitionID] = &result.LatestBlockInfo
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor encountered an error: %w", err)
	}

	return blockMap, nil
}

func (s *MongoBillStore) GetBlocksInRange(ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int) ([]*api.BlockInfo, uint64, error) {
	filter := bson.M{
		partitionIDKey: partitionID,
		blockNumberKey: bson.M{"$lte": dbStartBlock},
	}
	opts := options.Find().
		SetSort(bson.D{{Key: blockNumberKey, Value: -1}}).
		SetLimit(int64(count))

	cursor, err := s.db.Collection(blocksCollectionName).Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query blocks: %w", err)
	}
	defer cursor.Close(ctx)

	var blocks []*api.BlockInfo
	for cursor.Next(ctx) {
		var block api.BlockInfo
		if err = cursor.Decode(&block); err != nil {
			return nil, 0, fmt.Errorf("failed to decode block: %w", err)
		}
		blocks = append(blocks, &block)
	}

	if len(blocks) == 0 {
		return nil, 0, fmt.Errorf("no blocks found for partitionID: %d", partitionID)
	}

	var prevBlockNumber uint64
	if len(blocks) == count {
		prevBlockNumber = blocks[len(blocks)-1].BlockNumber - 1
	} else {
		prevBlockNumber = 0
	}

	return blocks, prevBlockNumber, nil
}
