package main

import (
	exTypes "github.com/alphabill-org/alphabill-explorer-backend/types"
)

type BillStore interface {
	GetBlockNumber() (uint64, error)
	SetBlockNumber(blockNumber uint64) error

	// bill_store/blocks.go
	SetBlockInfo(b *exTypes.BlockInfo) error
	GetLastBlockNumber() (uint64, error)
	GetBlockInfo(blockNumber uint64) (*exTypes.BlockInfo, error)
	GetBlocksInfo(dbStartBlock uint64, count int) (res []*exTypes.BlockInfo, prevBlockNumber uint64, err error)

	// bill_store/txs.go
	SetTxInfo(txInfo *exTypes.TxInfo) error
	GetTxInfo(txHash string) (*exTypes.TxInfo, error)
	GetBlockTxsByBlockNumber(blockNumber uint64) (res []*exTypes.TxInfo, err error)

	// bill_store/units.go
	SetUnitID(unitID string, txHash string) error
}
