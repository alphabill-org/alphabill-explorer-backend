package types

import (
	"github.com/alphabill-org/alphabill/types"
)

type BillStore interface {
	GetBlockNumber() (uint64, error)
	SetBlockNumber(blockNumber uint64) error

	// bill_store/blocks.go
	SetBlockInfo(b *BlockInfo) error
	GetLastBlockNumber() (uint64, error)
	GetBlockInfo(blockNumber uint64) (*BlockInfo, error)
	GetBlocksInfo(dbStartBlock uint64, count int) (res []*BlockInfo, prevBlockNumber uint64, err error)

	// bill_store/txs.go
	SetTxInfo(txInfo *TxInfo) error
	GetTxInfo(txHash string) (*TxInfo, error)
	GetBlockTxsByBlockNumber(blockNumber uint64) (res []*TxInfo, err error)

	// bill_store/units.go
	SetUnit(unit types.UnitID, txHash []byte) error
}
