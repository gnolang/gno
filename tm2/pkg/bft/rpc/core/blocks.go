package core

import (
	"fmt"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// BlockchainInfo gets block headers for minHeight <= height <= maxHeight.
// Block headers are returned in descending order (highest first).
//
// Returns at most 20 items.
func (env *Environment) BlockchainInfo(ctx *rpctypes.Context, minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "BlockchainInfo")
	defer span.End()
	// maximum 20 block metas
	const limit int64 = 20
	var err error
	minHeight, maxHeight, err = filterMinMax(env.BlockStore.Height(), minHeight, maxHeight, limit)
	if err != nil {
		return nil, err
	}
	env.Logger.Debug("BlockchainInfoHandler", "maxHeight", maxHeight, "minHeight", minHeight)

	blockMetas := []*types.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := env.BlockStore.LoadBlockMeta(height)
		if blockMeta == nil {
			return nil, fmt.Errorf("block meta not found for height %d", height)
		}
		blockMetas = append(blockMetas, blockMeta)
	}

	return &ctypes.ResultBlockchainInfo{
		LastHeight: env.BlockStore.Height(),
		BlockMetas: blockMetas,
	}, nil
}

// error if either low or high are negative or low > high
// if low is 0 it defaults to 1, if high is 0 it defaults to height (block height).
// limit sets the maximum amounts of values included within [low,high] (inclusive),
// increasing low as necessary.
func filterMinMax(height, low, high, limit int64) (int64, int64, error) {
	// filter negatives
	if low < 0 || high < 0 {
		return low, high, fmt.Errorf("heights must be non-negative")
	}

	// adjust for default values
	if low == 0 {
		low = 1
	}
	if high == 0 {
		high = height
	}

	// limit high to the height
	high = min(height, high)

	// limit low to within `limit` of max
	// so the total number of blocks returned will be `limit`
	low = max(low, high-limit+1)

	if low > high {
		return low, high, fmt.Errorf("min height %d can't be greater than max height %d", low, high)
	}
	return low, high, nil
}

// Block returns the block at the given height. If no height is provided, it
// fetches the latest block.
func (env *Environment) Block(ctx *rpctypes.Context, heightPtr *int64) (*ctypes.ResultBlock, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "Block")
	defer span.End()
	storeHeight := env.BlockStore.Height()
	height, err := getHeight(storeHeight, heightPtr)
	if err != nil {
		return nil, err
	}

	blockMeta := env.BlockStore.LoadBlockMeta(height)
	if blockMeta == nil {
		return nil, fmt.Errorf("block meta not found for height %d", height)
	}
	block := env.BlockStore.LoadBlock(height)
	if block == nil {
		return nil, fmt.Errorf("block not found for height %d", height)
	}
	return &ctypes.ResultBlock{BlockMeta: blockMeta, Block: block}, nil
}

// Commit returns the block commit at the given height. If no height is
// provided, it fetches the commit for the latest block.
func (env *Environment) Commit(ctx *rpctypes.Context, heightPtr *int64) (*ctypes.ResultCommit, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "Commit")
	defer span.End()
	storeHeight := env.BlockStore.Height()
	height, err := getHeight(storeHeight, heightPtr)
	if err != nil {
		return nil, err
	}

	blockMeta := env.BlockStore.LoadBlockMeta(height)
	if blockMeta == nil {
		return nil, fmt.Errorf("block meta not found for height %d", height)
	}
	header := blockMeta.Header

	// If the next block has not been committed yet,
	// use a non-canonical commit
	if height == storeHeight {
		commit := env.BlockStore.LoadSeenCommit(height)
		if commit == nil {
			return nil, fmt.Errorf("seen commit not found for height %d", height)
		}
		return ctypes.NewResultCommit(&header, commit, false), nil
	}

	// Return the canonical commit (comes from the block at height+1)
	commit := env.BlockStore.LoadBlockCommit(height)
	if commit == nil {
		return nil, fmt.Errorf("block commit not found for height %d", height)
	}
	return ctypes.NewResultCommit(&header, commit, true), nil
}

// BlockResults gets ABCIResults at a given height. If no height is provided,
// it fetches results for the latest block. Results are for the height of the
// block containing the txs.
func (env *Environment) BlockResults(ctx *rpctypes.Context, heightPtr *int64) (*ctypes.ResultBlockResults, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "BlockResults")
	defer span.End()
	storeHeight := env.BlockStore.Height()
	height, err := getHeightWithMin(storeHeight, heightPtr, 0)
	if err != nil {
		return nil, err
	}

	results, err := sm.LoadABCIResponses(env.StateDB, height)
	if err != nil {
		return nil, err
	}

	res := &ctypes.ResultBlockResults{
		Height:  height,
		Results: results,
	}
	return res, nil
}

func getHeight(currentHeight int64, heightPtr *int64) (int64, error) {
	return getHeightWithMin(currentHeight, heightPtr, 1)
}

func getHeightWithMin(currentHeight int64, heightPtr *int64, minVal int64) (int64, error) {
	if heightPtr != nil {
		height := *heightPtr
		if height < minVal {
			return 0, fmt.Errorf("height must be greater than or equal to %d", minVal)
		}
		if height > currentHeight {
			return 0, fmt.Errorf("height must be less than or equal to the current blockchain height")
		}
		return height, nil
	}
	return currentHeight, nil
}
