package tx

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/params"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Handler is the transaction RPC handler
type Handler struct {
	stateDB    dbm.DB
	blockStore sm.BlockStore
}

// NewHandler creates a new instance of the transaction RPC handler
func NewHandler(blockStore sm.BlockStore, stateDB dbm.DB) *Handler {
	return &Handler{
		blockStore: blockStore,
		stateDB:    stateDB,
	}
}

// TxHandler allows for querying the transaction results. `nil` could mean the
// transaction is in the mempool, invalidated, or was not sent in the first
// place.
//
//	Params:
//	- hash []byte (required)
func (h *Handler) TxHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	_, span := traces.Tracer().Start(context.Background(), "Tx")
	defer span.End()

	const idxHash = 0

	hash, err := params.AsBytes(p, idxHash, true)
	if err != nil {
		return nil, err
	}

	// Get the result index from storage, if any
	resultIndex, loadIdxErr := sm.LoadTxResultIndex(h.stateDB, hash)
	if loadIdxErr != nil {
		return nil, spec.GenerateResponseError(loadIdxErr)
	}

	storeHeight := h.blockStore.Height()
	if resultIndex.BlockNum < 1 || resultIndex.BlockNum > storeHeight {
		return nil, spec.GenerateResponseError(
			fmt.Errorf(
				"height (%d) must be less than or equal to the current blockchain height (%d)",
				resultIndex.BlockNum,
				storeHeight,
			),
		)
	}

	// Load the block
	block := h.blockStore.LoadBlock(resultIndex.BlockNum)
	if block == nil {
		return nil, spec.GenerateResponseError(
			fmt.Errorf("block not found for height %d", resultIndex.BlockNum),
		)
	}

	numTxs := len(block.Txs)
	if numTxs == 0 || int(resultIndex.TxIndex) >= numTxs {
		return nil, spec.GenerateResponseError(
			fmt.Errorf(
				"unable to get block transaction for block %d, index %d",
				resultIndex.BlockNum,
				resultIndex.TxIndex,
			),
		)
	}

	rawTx := block.Txs[resultIndex.TxIndex]

	// Fetch the block results
	blockResults, loadResErr := sm.LoadABCIResponses(h.stateDB, resultIndex.BlockNum)
	if loadResErr != nil {
		return nil, spec.GenerateResponseError(
			fmt.Errorf("unable to load block results, %w", loadResErr),
		)
	}

	if int(resultIndex.TxIndex) >= len(blockResults.DeliverTxs) {
		return nil, spec.GenerateResponseError(
			fmt.Errorf(
				"unable to get deliver result for block %d, index %d",
				resultIndex.BlockNum,
				resultIndex.TxIndex,
			),
		)
	}

	deliverResponse := blockResults.DeliverTxs[resultIndex.TxIndex]

	return &ResultTx{
		Hash:     hash,
		Height:   resultIndex.BlockNum,
		Index:    resultIndex.TxIndex,
		TxResult: deliverResponse,
		Tx:       rawTx,
	}, nil
}
