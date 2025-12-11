package mempool

import (
	"fmt"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	coreparams "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/params"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/utils"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
)

// Handler is the mempool RPC handler
type Handler struct {
	mempool    Mempool
	dispatcher *txDispatcher
}

// NewHandler creates a new instance of the mempool RPC handler.
func NewHandler(
	mp Mempool,
	evsw events.EventSwitch,
	timeoutBroadcastTxCommit time.Duration, // TODO use config?
) *Handler {
	return &Handler{
		mempool:    mp,
		dispatcher: newTxDispatcher(evsw, timeoutBroadcastTxCommit),
	}
}

// BroadcastTxAsyncHandler broadcasts the tx and returns right away, with no response.
// Does not wait for CheckTx nor DeliverTx results
//
//		Params:
//	  - tx   []byte (required)
func (h *Handler) BroadcastTxAsyncHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	const idxTx = 0

	rawTx, err := coreparams.AsBytes(p, idxTx, true)
	if err != nil {
		return nil, err
	}

	tx := types.Tx(rawTx)

	if checkErr := h.mempool.CheckTx(tx, nil); checkErr != nil {
		return nil, spec.GenerateResponseError(checkErr)
	}

	return &ctypes.ResultBroadcastTx{
		Hash: tx.Hash(),
	}, nil
}

// BroadcastTxSyncHandler broadcasts the tx and returns with the response from CheckTx.
// Does not wait for DeliverTx result
//
//		Params:
//	  - tx   []byte (required)
func (h *Handler) BroadcastTxSyncHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	const idxTx = 0

	rawTx, err := coreparams.AsBytes(p, idxTx, true)
	if err != nil {
		return nil, err
	}

	tx := types.Tx(rawTx)

	resCh := make(chan abci.Response, 1)
	if checkErr := h.mempool.CheckTx(tx, func(res abci.Response) {
		resCh <- res
	}); checkErr != nil {
		return nil, spec.GenerateResponseError(checkErr)
	}

	res := <-resCh
	r := res.(abci.ResponseCheckTx)

	return &ctypes.ResultBroadcastTx{
		Error: r.Error,
		Data:  r.Data,
		Log:   r.Log,
		Hash:  tx.Hash(),
	}, nil
}

// BroadcastTxCommitHandler broadcasts the tx and returns with the responses from CheckTx and DeliverTx.
//
//		Params:
//	  - tx   []byte (required)
func (h *Handler) BroadcastTxCommitHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	const idxTx = 0

	rawTx, err := coreparams.AsBytes(p, idxTx, true)
	if err != nil {
		return nil, err
	}

	tx := types.Tx(rawTx)

	checkTxResCh := make(chan abci.Response, 1)
	if checkErr := h.mempool.CheckTx(tx, func(res abci.Response) {
		checkTxResCh <- res
	}); checkErr != nil {
		return nil, spec.GenerateResponseError(
			fmt.Errorf("error on BroadcastTxCommit: %w", checkErr),
		)
	}

	checkTxResMsg := <-checkTxResCh
	checkTxRes := checkTxResMsg.(abci.ResponseCheckTx)

	if checkTxRes.Error != nil {
		return &ctypes.ResultBroadcastTxCommit{
			CheckTx:   checkTxRes,
			DeliverTx: abci.ResponseDeliverTx{},
			Hash:      tx.Hash(),
		}, nil
	}

	txRes, txErr := h.dispatcher.getTxResult(tx, nil)
	if txErr != nil {
		return nil, spec.GenerateResponseError(txErr)
	}

	return &ctypes.ResultBroadcastTxCommit{
		CheckTx:   checkTxRes,
		DeliverTx: txRes.Response,
		Hash:      tx.Hash(),
		Height:    txRes.Height,
	}, nil
}

// UnconfirmedTxsHandler fetches unconfirmed transactions (maximum ?limit entries) including their number.
//
//		Params:
//	  - limit	int64 (optional, default 30, max 100)
func (h *Handler) UnconfirmedTxsHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	const idxLimit = 0

	limit64, err := coreparams.AsInt64(p, idxLimit)
	if err != nil {
		return nil, err
	}

	var (
		limit = utils.ValidatePerPage(int(limit64))
		txs   = h.mempool.ReapMaxTxs(limit)
	)

	return &ctypes.ResultUnconfirmedTxs{
		Count:      len(txs),
		Total:      h.mempool.Size(),
		TotalBytes: h.mempool.TxsBytes(),
		Txs:        txs,
	}, nil
}

// NumUnconfirmedTxsHandler fetches the number of unconfirmed transactions.
//
//	No params
func (h *Handler) NumUnconfirmedTxsHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	if len(p) > 0 {
		return nil, spec.GenerateInvalidParamError(1)
	}

	return &ctypes.ResultUnconfirmedTxs{
		Count:      h.mempool.Size(),
		Total:      h.mempool.Size(),
		TotalBytes: h.mempool.TxsBytes(),
	}, nil
}
