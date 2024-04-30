package blocks

import (
	"context"
	"crypto"
	"fmt"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	abtypes "github.com/alphabill-org/alphabill/types"
)

const (
	DustBillDeletionTimeout    = 65536
	ExpiredBillDeletionTimeout = 65536
)

type Store interface {
	GetBlockNumber() (uint64, error)
	SetBlockNumber(blockNumber uint64) error
	SetTxInfo(txExplorer *api.TxInfo) error
	SetBlockInfo(b *api.BlockInfo) error
}

type BlockProcessor struct {
	store Store
}

func NewBlockProcessor(store Store, moneySystemID abtypes.SystemID) (*BlockProcessor, error) {
	return &BlockProcessor{store: store}, nil
}

func (p *BlockProcessor) ProcessBlock(_ context.Context, b *abtypes.Block) error {
	roundNumber := b.GetRoundNumber()
	//println("processing block: ", roundNumber)
	if len(b.Transactions) > 0 {
		fmt.Printf("Block number: %d has %d transactions\n", roundNumber, len(b.Transactions))
	}
	lastBlockNumber, err := p.store.GetBlockNumber()
	if err != nil {
		return err
	}
	if lastBlockNumber >= roundNumber {
		return fmt.Errorf("invalid block number. Received blockNumber %d current wallet blockNumber %d", roundNumber, lastBlockNumber)
	}
	for i, tx := range b.Transactions {

		if err := p.processTx(tx, b, i); err != nil {
			return fmt.Errorf("failed to process transaction: %w", err)
		}

		txInfo, err := api.NewTxInfo(roundNumber, tx)

		if err != nil {
			return fmt.Errorf("failed create new txInfo in ProcessBlock: %w", err)
		}

		err = p.saveTx(txInfo)
		if err != nil {
			return fmt.Errorf("failed to save tx in ProcessBlock: %w", err)
		}
	}
	err = p.saveBlock(b)
	if err != nil {
		return err
	}
	//fmt.Printf("roundNumber: %d	, count: %d 	", roundNumber , len(b.Transactions))
	return p.store.SetBlockNumber(roundNumber)
}

func (p *BlockProcessor) processTx(txr *abtypes.TransactionRecord, b *abtypes.Block, txIdx int) error {
	txo := txr.TransactionOrder
	txHash := txo.Hash(crypto.SHA256)
	_ = txHash
	proof, _, err := abtypes.NewTxProof(b, txIdx, crypto.SHA256)
	if err != nil {
		return err
	}

	_ = proof

	//switch txo.PayloadType() {
	//case moneytx.PayloadTypeTransfer:
	//	println(fmt.Sprintf("received transfer order (UnitID=%x)", txo.UnitID()))
	//	//if err = p.updateFCB(dbTx, txr); err != nil {
	//	//	return err
	//	//}
	//	attr := &moneytx.TransferAttributes{}
	//	if err = txo.UnmarshalAttributes(attr); err != nil {
	//		return err
	//	}
	//	if err = dbTx.SetBill(&types.Bill{
	//		Id:             txo.UnitID(),
	//		Value:          attr.TargetValue,
	//		TxHash:         txHash,
	//		OwnerPredicate: attr.NewBearer,
	//	}, proof); err != nil {
	//		return err
	//	}
	//	if err = saveTx(dbTx, attr.NewBearer, txo, txHash); err != nil {
	//		return err
	//	}
	//case moneytx.PayloadTypeTransDC:
	//	println(fmt.Sprintf("received TransferDC order (UnitID=%x)", txo.UnitID()))
	//	//err := p.updateFCB(dbTx, txr)
	//	//if err != nil {
	//	//	return err
	//	//}
	//	attr := &moneytx.TransferDCAttributes{}
	//	err = txo.UnmarshalAttributes(attr)
	//	if err != nil {
	//		return err
	//	}
	//
	//	// update bill value, txHash, target unit
	//	dcBill, err := dbTx.GetBill(txo.UnitID())
	//	if err != nil {
	//		return fmt.Errorf("failed to fetch bill: %w", err)
	//	}
	//	if dcBill == nil {
	//		return fmt.Errorf("bill not found: %x", txo.UnitID())
	//	}
	//	dcBill.Value = attr.Value
	//	dcBill.TxHash = txHash
	//	dcBill.DCTargetUnitID = attr.TargetUnitID
	//	//dcBill.DCTargetUnitBacklink = attr.TargetUnitBacklink
	//	err = dbTx.SetBill(dcBill, proof)
	//	if err != nil {
	//		return err
	//	}
	//	// TODO AB-1133
	//	//err = dbTx.SetBillExpirationTime(roundNumber+DustBillDeletionTimeout, txo.UnitID())
	//	//if err != nil {
	//	//	return err
	//	//}
	//case moneytx.PayloadTypeSplit:
	//	//err := p.updateFCB(dbTx, txr)
	//	//if err != nil {
	//	//	return err
	//	//}
	//	attr := &moneytx.SplitAttributes{}
	//	err = txo.UnmarshalAttributes(attr)
	//	if err != nil {
	//		return err
	//	}
	//	// old bill
	//	oldBill, err := dbTx.GetBill(txo.UnitID())
	//	if err != nil {
	//		return err
	//	}
	//	if oldBill != nil {
	//		println(fmt.Sprintf("received split order (existing UnitID=%x)", txo.UnitID()))
	//		err = dbTx.SetBill(&types.Bill{
	//			Id:             txo.UnitID(),
	//			Value:          attr.RemainingValue,
	//			TxHash:         txHash,
	//			OwnerPredicate: oldBill.OwnerPredicate,
	//		}, proof)
	//		if err != nil {
	//			return err
	//		}
	//	} else {
	//		// we should always have the "previous bill" other than splitting the initial bill or some error condition
	//		println(fmt.Sprintf("received split order where existing unit was not found, ignoring tx (unitID=%x)", txo.UnitID()))
	//	}
	//
	//	// new bill
	//	newID := moneytx.NewBillID(txo.UnitID(), moneytx.HashForIDCalculation(txo.UnitID(), txo.Payload.Attributes, txo.Timeout(), 0, crypto.SHA256)) // TODO fix
	//	println(fmt.Sprintf("received split order (new UnitID=%x)", newID))
	//	err = dbTx.SetBill(&types.Bill{
	//		Id:     newID,
	//		Value:  0, // attr.Amount, TODO fix
	//		TxHash: txHash,
	//		//OwnerPredicate: attr.TargetBearer, // TODO
	//	}, proof)
	//	if err != nil {
	//		return err
	//	}
	//	if err = saveTx(dbTx, nil, txo, txHash); err != nil { // TODO fix
	//		return err
	//	}
	//case moneytx.PayloadTypeSwapDC:
	//	//err := p.updateFCB(dbTx, txr)
	//	//if err != nil {
	//	//	return err
	//	//}
	//	attr := &moneytx.SwapDCAttributes{}
	//	err = txo.UnmarshalAttributes(attr)
	//	if err != nil {
	//		return err
	//	}
	//	bill, err := dbTx.GetBill(txo.UnitID())
	//	if err != nil {
	//		return err
	//	}
	//	if bill == nil {
	//		return fmt.Errorf("existing bill not found for swap tx (UnitID=%x)", txo.UnitID())
	//	}
	//	println(fmt.Sprintf("received swap order (UnitID=%x)", txo.UnitID()))
	//	bill.Value += attr.TargetValue
	//	bill.TxHash = txHash
	//	bill.OwnerPredicate = attr.OwnerCondition
	//	err = dbTx.SetBill(bill, proof)
	//	if err != nil {
	//		return err
	//	}
	//	for _, dustTransfer := range attr.DcTransfers {
	//		err := dbTx.RemoveBill(dustTransfer.TransactionOrder.UnitID())
	//		if err != nil {
	//			return err
	//		}
	//	}
	//case transactions.PayloadTypeTransferFeeCredit:
	//	println(fmt.Sprintf("received transferFC order (UnitID=%x), hash: '%X'", txo.UnitID(), txHash))
	//	bill, err := dbTx.GetBill(txo.UnitID())
	//	if err != nil {
	//		return fmt.Errorf("failed to get bill: %w", err)
	//	}
	//	if bill == nil {
	//		return fmt.Errorf("unit not found for transferFC tx (unitID=%X)", txo.UnitID())
	//	}
	//	attr := &transactions.TransferFeeCreditAttributes{}
	//	err = txo.UnmarshalAttributes(attr)
	//	if err != nil {
	//		return fmt.Errorf("failed to unmarshal transferFC attributes: %w", err)
	//	}
	//	if attr.Amount < bill.Value {
	//		bill.Value -= attr.Amount
	//		bill.TxHash = txHash
	//		if err := dbTx.SetBill(bill, proof); err != nil {
	//			return fmt.Errorf("failed to save transferFC bill with proof: %w", err)
	//		}
	//	} else {
	//		if err := dbTx.StoreTxProof(txo.UnitID(), txHash, proof); err != nil {
	//			return fmt.Errorf("failed to types tx proof zero value bill: %w", err)
	//		}
	//		if err := dbTx.RemoveBill(bill.Id); err != nil {
	//			return fmt.Errorf("failed to remove zero value bill: %w", err)
	//		}
	//	}
	//	err = p.addTransferredCreditToPartitionFeeBill(dbTx, attr, proof, txr.ServerMetadata.ActualFee)
	//	if err != nil {
	//		return fmt.Errorf("failed to add transferred fee credit to partition fee bill: %w", err)
	//	}
	//	err = p.addTxFeeToMoneyFeeBill(dbTx, txr, proof)
	//	if err != nil {
	//		return fmt.Errorf("failed to add tx fees to money fee bill: %w", err)
	//	}
	//	err = p.addLockedFeeCredit(dbTx, attr.TargetSystemIdentifier, attr.TargetRecordID, txr)
	//	if err != nil {
	//		return fmt.Errorf("failed to add locked fee credit: %w", err)
	//	}
	//	return nil
	//case transactions.PayloadTypeAddFeeCredit:
	//	println(fmt.Sprintf("received addFC order (UnitID=%x), hash: '%X'", txo.UnitID(), txHash))
	//	fcb, err := dbTx.GetFeeCreditBill(txo.UnitID())
	//	if err != nil {
	//		return err
	//	}
	//	addFCAttr := &transactions.AddFeeCreditAttributes{}
	//	err = txo.UnmarshalAttributes(addFCAttr)
	//	if err != nil {
	//		return err
	//	}
	//	transferFCAttr := &transactions.TransferFeeCreditAttributes{}
	//	err = addFCAttr.FeeCreditTransfer.TransactionOrder.UnmarshalAttributes(transferFCAttr)
	//	if err != nil {
	//		return err
	//	}
	//	return dbTx.SetFeeCreditBill(&Bill{
	//		Id:              txo.UnitID(),
	//		Value:           fcb.getValue() + transferFCAttr.Amount - addFCAttr.FeeCreditTransfer.ServerMetadata.ActualFee - txr.ServerMetadata.ActualFee,
	//		TxHash:          txHash,
	//		LastAddFCTxHash: txHash,
	//	}, proof)
	//case transactions.PayloadTypeCloseFeeCredit:
	//	println(fmt.Sprintf("received closeFC order (UnitID=%x)", txo.UnitID()))
	//	fcb, err := dbTx.GetFeeCreditBill(txo.UnitID())
	//	if err != nil {
	//		return err
	//	}
	//	attr := &transactions.CloseFeeCreditAttributes{}
	//	err = txo.UnmarshalAttributes(attr)
	//	if err != nil {
	//		return err
	//	}
	//	err = p.addClosedFeeCredit(dbTx, txo.UnitID(), txr)
	//	if err != nil {
	//		return err
	//	}
	//	return dbTx.SetFeeCreditBill(&Bill{
	//		Id:              txo.UnitID(),
	//		TxHash:          txHash,
	//		Value:           fcb.getValue() - attr.Amount,
	//		LastAddFCTxHash: fcb.getLastAddFCTxHash(),
	//	}, proof)
	//case transactions.PayloadTypeReclaimFeeCredit:
	//	println(fmt.Sprintf("received reclaimFC order (UnitID=%x)", txo.UnitID()))
	//	bill, err := dbTx.GetBill(txo.UnitID())
	//	if err != nil {
	//		return err
	//	}
	//	if bill == nil {
	//		return fmt.Errorf("unit not found for reclaimFC tx (unitID=%X)", txo.UnitID())
	//	}
	//	reclaimFCAttr := &transactions.ReclaimFeeCreditAttributes{}
	//	err = txo.UnmarshalAttributes(reclaimFCAttr)
	//	if err != nil {
	//		return err
	//	}
	//	closeFCTXR := reclaimFCAttr.CloseFeeCreditTransfer
	//	closeFCTXO := closeFCTXR.TransactionOrder
	//	closeFCAttr := &transactions.CloseFeeCreditAttributes{}
	//	err = closeFCTXO.UnmarshalAttributes(closeFCAttr)
	//	if err != nil {
	//		return err
	//	}
	//
	//	// 1. remove reclaimed amount from user bill
	//	reclaimedValue := closeFCAttr.Amount - closeFCTXR.ServerMetadata.ActualFee - txr.ServerMetadata.ActualFee
	//	bill.Value += reclaimedValue
	//	bill.TxHash = txHash
	//	err = dbTx.SetBill(bill, proof)
	//	if err != nil {
	//		return err
	//	}
	//	// 2. remove reclaimed amount from partition fee bill
	//	//err = p.removeReclaimedCreditFromPartitionFeeBill(dbTx, closeFCTXR, closeFCAttr, proof)
	//	//if err != nil {
	//	//	return err
	//	//}
	//	//// 3. add reclaimFC tx fee to money partition fee bill
	//	//return p.addTxFeeToMoneyFeeBill(dbTx, txr, proof)
	//default:
	//	println(fmt.Sprintf("received unknown transaction type, skipping processing: %s", txo.PayloadType()))
	//	return nil
	//}
	return nil
}

func (p *BlockProcessor) saveTx(txInfo *api.TxInfo) error {
	if txInfo == nil {
		return fmt.Errorf("transaction is nil")
	}
	err := p.store.SetTxInfo(txInfo)
	if err != nil {
		return err
	}
	return nil
}

func (p *BlockProcessor) saveBlock(b *abtypes.Block) error {
	if b == nil {
		return fmt.Errorf("block is nil")
	}
	blockInfo, err := api.NewBlockInfo(b)
	if err != nil {
		return err
	}
	err = p.store.SetBlockInfo(blockInfo)
	if err != nil {
		return err
	}
	return nil
}
