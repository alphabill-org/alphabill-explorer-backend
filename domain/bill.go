package domain

import "github.com/alphabill-org/alphabill-go-base/types"

type Bill struct {
	NetworkID   types.NetworkID
	PartitionID types.PartitionID
	ID          types.UnitID
	Value       uint64
	LockStatus  uint64
	Counter     uint64
}
