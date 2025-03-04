package rpc

import (
	"context"

	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-wallet/client/rpc"
	sdktypes "github.com/alphabill-org/alphabill-wallet/client/types"
)

type StateAPIClient struct {
	*rpc.StateAPIClient
}

func NewStateAPIClient(ctx context.Context, url string) (*StateAPIClient, error) {
	client, err := rpc.NewStateAPIClient(ctx, url)
	if err != nil {
		return nil, err
	}

	return &StateAPIClient{client}, nil
}

func (c *StateAPIClient) GetUnit(ctx context.Context, unitID types.UnitID, includeStateProof bool) (*sdktypes.Unit[any], error) {
	var res *sdktypes.Unit[any]
	err := c.RpcClient.CallContext(ctx, &res, "state_getUnit", unitID, includeStateProof)
	return res, err
}
