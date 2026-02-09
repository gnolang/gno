package blocks

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/params"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/utils"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Handler is the blocks RPC handler
type Handler struct {
	store   state.BlockStore
	stateDB dbm.DB
}

// NewHandler creates a new instance of the blocks RPC handler
func NewHandler(store state.BlockStore, stateDB dbm.DB) *Handler {
	return &Handler{
		store:   store,
		stateDB: stateDB,
	}
}

// BlockchainInfoHandler fetches block headers for a given range.
// Block headers are returned in descending order (highest first)
//
//		Params:
//	  - minHeight   int64 (optional, default 1)
//	  - maxHeight   int64 (optional, default latest height)
func (h *Handler) BlockchainInfoHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	_, span := traces.Tracer().Start(context.Background(), "BlockchainInfo")
	defer span.End()

	const limit int64 = 20

	const (
		idxMinHeight = 0
		idxMaxHeight = 1
	)

	minHeight, err := params.AsInt64(p, idxMinHeight)
	if err != nil {
		return nil, err
	}

	maxHeight, err := params.AsInt64(p, idxMaxHeight)
	if err != nil {
		return nil, err
	}

	// Grab the latest height
	storeHeight := h.store.Height()

	minHeight, maxHeight, filterErr := filterMinMax(storeHeight, minHeight, maxHeight, limit)
	if filterErr != nil {
		return nil, spec.GenerateResponseError(filterErr)
	}

	blockMetas := make([]*types.BlockMeta, 0, maxHeight-minHeight+1)
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := h.store.LoadBlockMeta(height)

		if blockMeta == nil {
			// This would be a huge problemo
			continue
		}

		blockMetas = append(blockMetas, blockMeta)
	}

	return &ResultBlockchainInfo{
		LastHeight: storeHeight,
		BlockMetas: blockMetas,
	}, nil
}

// BlockHandler fetches the block at the given height.
// If no height is provided, it will fetch the latest block
//
//		Params:
//	  - height   int64 (optional, default latest height)
func (h *Handler) BlockHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	_, span := traces.Tracer().Start(context.Background(), "Block")
	defer span.End()

	const idxHeight = 0

	storeHeight := h.store.Height()

	height, err := params.AsInt64(p, idxHeight)
	if err != nil {
		return nil, err
	}

	height, normalizeErr := utils.NormalizeHeight(storeHeight, height, 1)
	if normalizeErr != nil {
		return nil, spec.GenerateResponseError(normalizeErr)
	}

	blockMeta := h.store.LoadBlockMeta(height)
	if blockMeta == nil {
		return nil, spec.GenerateResponseError(
			fmt.Errorf("block meta not found for height %d", height),
		)
	}

	block := h.store.LoadBlock(height)
	if block == nil {
		return nil, spec.GenerateResponseError(
			fmt.Errorf("block not found for height %d", height),
		)
	}

	return &ResultBlock{
		BlockMeta: blockMeta,
		Block:     block,
	}, nil
}

// CommitHandler fetches the block commit for the given height.
// If no height is provided, it will fetch the commit for the latest block
//
//		Params:
//	  - height   int64 (optional, default latest height)
func (h *Handler) CommitHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	_, span := traces.Tracer().Start(context.Background(), "Commit")
	defer span.End()

	const idxHeight = 0

	storeHeight := h.store.Height()

	height, err := params.AsInt64(p, idxHeight)
	if err != nil {
		return nil, err
	}

	height, normalizeErr := utils.NormalizeHeight(storeHeight, height, 1)
	if normalizeErr != nil {
		return nil, spec.GenerateResponseError(normalizeErr)
	}

	blockMeta := h.store.LoadBlockMeta(height)
	if blockMeta == nil {
		return nil, spec.GenerateResponseError(
			fmt.Errorf("block meta not found for height %d", height),
		)
	}

	header := blockMeta.Header

	if height == storeHeight {
		// latest, non-canonical commit
		commit := h.store.LoadSeenCommit(height)
		if commit == nil {
			return nil, spec.GenerateResponseError(
				fmt.Errorf("seen commit not found for height %d", height),
			)
		}

		return NewResultCommit(&header, commit, false), nil
	}

	// canonical commit (from height+1)
	commit := h.store.LoadBlockCommit(height)
	if commit == nil {
		return nil, spec.GenerateResponseError(
			fmt.Errorf("canonical commit not found for height %d", height),
		)
	}

	return NewResultCommit(&header, commit, true), nil
}

// BlockResultsHandler fetches the ABCI results for the given height.
// If no height is provided, it will fetch results for the latest block
//
//		Params:
//	  - height   int64 (optional, default latest height)
func (h *Handler) BlockResultsHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	_, span := traces.Tracer().Start(context.Background(), "BlockResults")
	defer span.End()

	storeHeight := h.store.Height()

	height, err := params.AsInt64(p, 0)
	if err != nil {
		return nil, err
	}

	height, normalizeErr := utils.NormalizeHeight(storeHeight, height, 0)
	if normalizeErr != nil {
		return nil, spec.GenerateResponseError(normalizeErr)
	}

	results, loadErr := state.LoadABCIResponses(h.stateDB, height)
	if loadErr != nil {
		return nil, spec.GenerateResponseError(loadErr)
	}

	return &ResultBlockResults{
		Height:  height,
		Results: results,
	}, nil
}

// error if either low or high are negative or low > high
// if low is 0 it defaults to 1, if high is 0 it defaults to height (block height).
// limit sets the maximum amounts of values included within [low,high] (inclusive),
// increasing low as necessary.
// Migrated from legacy Tendermint RPC
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
