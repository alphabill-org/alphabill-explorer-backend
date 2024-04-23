package store

import (
	sdk "github.com/alphabill-org/alphabill-wallet/wallet"
	"github.com/alphabill-org/alphabill/network/protocol/genesis"
	"github.com/alphabill-org/alphabill/types"
)

type (
	// BillStore type for creating BillStoreTx transactions
	BillStore interface {
		Do() BillStoreTx
		WithTransaction(func(tx BillStoreTx) error) error
	}
	// BillStoreTx type for managing units by their ID and owner condition
	BillStoreTx interface {
		GetLastBlockNumber() (uint64, error)
		GetBlockByBlockNumber(blockNumber uint64) (*types.Block, error)
		GetBlocks(dbStartBlock uint64, count int) (res []*types.Block, prevBlockNumber uint64, err error)
		SetBlock(b *types.Block) error
		GetBlockExplorerByBlockNumber(blockNumber uint64) (*BlockExplorer, error)
		GetBlocksExplorer(dbStartBlock uint64, count int) (res []*BlockExplorer, prevBlockNumber uint64, err error)
		SetBlockExplorer(b *types.Block) error
		GetBlockExplorerTxsByBlockNumber(blockNumber uint64) (res []*TxExplorer, err error)
		GetBlockNumber() (uint64, error)
		SetBlockNumber(blockNumber uint64) error
		GetTxExplorerByTxHash(txHash string) (*TxExplorer, error)
		SetTxExplorerToBucket(txExplorer *TxExplorer) error
		GetBill(unitID []byte) (*Bill, error)
		GetBills(ownerCondition []byte, includeDCBills bool, offsetKey []byte, limit int) ([]*Bill, []byte, error)
		SetBill(bill *Bill, proof *types.TxProof) error
		RemoveBill(unitID []byte) error
		GetSystemDescriptionRecords() ([]*genesis.SystemDescriptionRecord, error)
		SetSystemDescriptionRecords(sdrs []*genesis.SystemDescriptionRecord) error
		GetTxProof(unitID types.UnitID, txHash sdk.TxHash) (*types.TxProof, error)
		//StoreTxProof(unitID types.UnitID, txHash sdk.TxHash, txProof *types.TxProof) error
		//GetTxHistoryRecords(dbStartKey []byte, count int) ([]*sdk.TxHistoryRecord, []byte, error)
		//GetTxHistoryRecordsByKey(hash sdk.PubKeyHash, dbStartKey []byte, count int) ([]*sdk.TxHistoryRecord, []byte, error)
		//StoreTxHistoryRecord(hash sdk.PubKeyHash, rec *sdk.TxHistoryRecord) error
	}

	Bills struct {
		Bills []*Bill `json:"bills"`
	}

	Bill struct {
		Id                   []byte `json:"id"`
		Value                uint64 `json:"value"`
		TxHash               []byte `json:"txHash"`
		DCTargetUnitID       []byte `json:"dcTargetUnitId,omitempty"`
		DCTargetUnitBacklink []byte `json:"dcTargetUnitBacklink,omitempty"`
		OwnerPredicate       []byte `json:"ownerPredicate"`

		// fcb specific fields
		// LastAddFCTxHash last add fee credit tx hash
		LastAddFCTxHash []byte `json:"lastAddFcTxHash,omitempty"`
	}

	BlockExplorer struct {
		_               struct{} `cbor:",toarray"`
		SystemID        *types.SystemID
		RoundNumber     uint64
		Header          *HeaderExplorer
		TxHashes        []string
		SummaryValue    []byte // summary value to certified
		SumOfEarnedFees uint64 // sum of the actual fees over all transaction records in the block
	}
	HeaderExplorer struct {
		_                 struct{} `cbor:",toarray"`
		Timestamp         uint64
		BlockHash         []byte
		PreviousBlockHash []byte
		ProposerID        string // validator
	}
	TxExplorer struct {
		_                struct{} `cbor:",toarray"`
		Hash             string
		BlockNumber      uint64
		Timeout          uint64
		PayloadType      string
		Status           *types.TxStatus
		TargetUnits      []types.UnitID
		TransactionOrder *types.TransactionOrder
		Fee              uint64
	}

	BlockInfo struct {
		_                  struct{} `cbor:",toarray"`
		Header             *types.Header
		TxHashes           []string
		UnicityCertificate *types.UnicityCertificate
	}
)
type PubKey []byte

type PubKeyHash []byte

type TxHash []byte

func (b *Bill) getTxHash() []byte {
	if b != nil {
		return b.TxHash
	}
	return nil
}

func (b *Bill) getValue() uint64 {
	if b != nil {
		return b.Value
	}
	return 0
}

func (b *Bill) getLastAddFCTxHash() []byte {
	if b != nil {
		return b.LastAddFCTxHash
	}
	return nil
}

func (b *Bill) IsDCBill() bool {
	if b != nil {
		return len(b.DCTargetUnitID) > 0
	}
	return false
}
