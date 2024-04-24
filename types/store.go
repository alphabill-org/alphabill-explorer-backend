package types

type BillStore interface {
	GetBlockNumber() (uint64, error)
	SetBlockNumber(blockNumber uint64) error
	SetTxInfo(txInfo *TxInfo) error
	GetTxInfo(txHash string) (*TxInfo, error)
	SetBlockInfo(b *BlockInfo) error
	GetLastBlockNumber() (uint64, error)
	GetBlockInfo(blockNumber uint64) (*BlockInfo, error)
	GetBlocksInfo(dbStartBlock uint64, count int) (res []*BlockInfo, prevBlockNumber uint64, err error)
}