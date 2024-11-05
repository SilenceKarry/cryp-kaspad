package kaspa

import (
	"context"
	"cryp-kaspad/internal/libs/kaspa/core/types"
	"fmt"

	"github.com/eoscanada/eos-go"
)

func (e *Eos) GetBlockNumberLatest(ctx context.Context) (int64, error) {
	result, err := e.Client.API.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("client.GetInfo error: %s", err)
	}

	return int64(result.LastIrreversibleBlockNum), nil
}

func (e *Eos) GetBlockByNumber(ctx context.Context, blockNumber int64) (*eos.BlockResp, error) {
	result, err := e.Client.API.GetBlockByNum(ctx, uint32(blockNumber))
	if err != nil {
		return nil, fmt.Errorf("client.GetBlockByNum error: %s", err)
	}

	return result, nil
}

func (e *Eos) GetBlockByNumberCustom(ctx context.Context, blockNumber uint32) (*types.Block, error) {
	result, err := e.Client.GetBlock(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("client.GetBlockByNum error: %s", err)
	}

	return result, nil
}
