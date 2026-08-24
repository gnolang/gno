package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// -----------------------------------------------------------------------------
// NOTE: tx should be signed, but this is only checked at the app level (not by Tendermint!)

// BroadcastTxAsync returns right away, with no response. Does not wait for
// CheckTx nor DeliverTx results.
//
// If you want to be sure that the transaction is included in a block, you can
// subscribe for the result using JSONRPC via a websocket. See
// https://docs.tendermint.com/v0.34/tendermint-core/subscription.html
func (env *Environment) BroadcastTxAsync(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "BroadcastTxAsync")
	defer span.End()
	err := env.Mempool.CheckTx(tx, nil)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultBroadcastTx{Hash: tx.Hash()}, nil
}

// BroadcastTxSync returns with the response from CheckTx. Does not wait for
// DeliverTx result.
func (env *Environment) BroadcastTxSync(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "BroadcastTxSync")
	defer span.End()
	resCh := make(chan abci.Response, 1)
	err := env.Mempool.CheckTx(tx, func(res abci.Response) {
		resCh <- res
	})
	if err != nil {
		return nil, err
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

// BroadcastTxCommit returns with the responses from CheckTx and DeliverTx.
//
// IMPORTANT: use only for testing and development. In production, use
// BroadcastTxSync or BroadcastTxAsync.
//
// CONTRACT: only returns error if mempool.CheckTx() errs, we timeout waiting
// for tx to commit, or the Environment was not started.
func (env *Environment) BroadcastTxCommit(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "BroadcastTxCommit")
	defer span.End()

	if env.txDispatcher == nil {
		return nil, errors.New("BroadcastTxCommit unavailable: Environment not started or no EventSwitch configured")
	}

	// Broadcast tx and wait for CheckTx result
	checkTxResCh := make(chan abci.Response, 1)
	err := env.Mempool.CheckTx(tx, func(res abci.Response) {
		checkTxResCh <- res
	})
	if err != nil {
		env.Logger.Error("Error on broadcastTxCommit", "err", err)
		return nil, fmt.Errorf("error on broadcastTxCommit: %w", err)
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

	// Wait for the tx to be included in a block or timeout.
	txRes, err := env.txDispatcher.getTxResult(tx, env.Config.TimeoutBroadcastTxCommit, nil)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultBroadcastTxCommit{
		CheckTx:   checkTxRes,
		DeliverTx: txRes.Response,
		Hash:      tx.Hash(),
		Height:    txRes.Height,
	}, nil
}

// UnconfirmedTxs gets unconfirmed transactions (maximum ?limit entries)
// including their number.
func (env *Environment) UnconfirmedTxs(ctx *rpctypes.Context, limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "UnconfirmedTxs")
	defer span.End()
	// reuse per_page validator
	limit = validatePerPage(limit)

	txs := env.Mempool.ReapMaxTxs(limit)
	return &ctypes.ResultUnconfirmedTxs{
		Count:      len(txs),
		Total:      env.Mempool.Size(),
		TotalBytes: env.Mempool.TxsBytes(),
		Txs:        txs,
	}, nil
}

// NumUnconfirmedTxs returns the number of unconfirmed transactions.
func (env *Environment) NumUnconfirmedTxs(ctx *rpctypes.Context) (*ctypes.ResultUnconfirmedTxs, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "NumUnconfirmedTxs")
	defer span.End()
	return &ctypes.ResultUnconfirmedTxs{
		Count:      env.Mempool.Size(),
		Total:      env.Mempool.Size(),
		TotalBytes: env.Mempool.TxsBytes(),
	}, nil
}

// ----------------------------------------
// txDispatcher

// NOTE: txDispatcher doesn't handle any throttling or resource management.
// The RPC websockets system is expected to throttle requests.
type txDispatcher struct {
	service.BaseService
	evsw       events.EventSwitch
	listenerID string
	sub        <-chan events.Event

	mtx     sync.Mutex
	waiters map[string]*txWaiter // string(types.Tx) -> *txWaiter
}

func newTxDispatcher(evsw events.EventSwitch) *txDispatcher {
	listenerID := fmt.Sprintf("txDispatcher#%v", random.RandStr(6))
	sub := events.SubscribeToEvent(evsw, listenerID, types.EventTx{})

	td := &txDispatcher{
		evsw:       evsw,
		listenerID: listenerID,
		sub:        sub,
		waiters:    make(map[string]*txWaiter),
	}
	td.BaseService = *service.NewBaseService(nil, "txDispatcher", td)
	err := td.Start()
	if err != nil {
		panic(err)
	}
	return td
}

func (td *txDispatcher) OnStart() error {
	go td.listenRoutine()
	return nil
}

func (td *txDispatcher) OnStop() {
	td.evsw.RemoveListener(td.listenerID)
}

func (td *txDispatcher) listenRoutine() {
	for {
		select {
		case event, ok := <-td.sub:
			if !ok {
				// The event switch closed our subscription during shutdown
				// (see events.SubscribeFilteredOn: the listener callback
				// closes the channel when it would otherwise block and
				// evsw.Quit() has fired). Stop cleanly rather than panic —
				// pending getTxResult waiters will time out normally.
				go func() { _ = td.Stop() }()
				return
			}
			txEvent := event.(types.EventTx)
			td.notifyTxEvent(txEvent)
		case <-td.Quit():
			return
		}
	}
}

func (td *txDispatcher) notifyTxEvent(txEvent types.EventTx) {
	td.mtx.Lock()
	defer td.mtx.Unlock()

	tx := txEvent.Result.Tx
	waiter, ok := td.waiters[string(tx)]
	if !ok {
		return // nothing to do
	}
	waiter.txRes = txEvent.Result
	close(waiter.waitCh)
}

// blocking
// If the tx is already being waited on, returns the result from the original request.
// Upon result or timeout, the tx is forgotten from txDispatcher, and can be re-requested.
// If the tx times out, an error is returned.
// Quit can optionally be provided to terminate early (e.g. if the caller disconnects).
func (td *txDispatcher) getTxResult(tx types.Tx, timeout time.Duration, quit chan struct{}) (types.TxResult, error) {
	// Get or create waiter.
	td.mtx.Lock()
	waiter, ok := td.waiters[string(tx)]
	if !ok {
		waiter = newTxWaiter(tx, timeout)
		td.waiters[string(tx)] = waiter
	}
	td.mtx.Unlock()

	select {
	case <-waiter.waitCh:
		return waiter.txRes, nil
	case <-waiter.timeCh:
		return types.TxResult{}, errors.New("request timeout")
	case <-quit:
		return types.TxResult{}, errors.New("caller quit")
	}
}

type txWaiter struct {
	tx     types.Tx
	waitCh chan struct{}
	timeCh <-chan time.Time
	txRes  types.TxResult
}

func newTxWaiter(tx types.Tx, timeout time.Duration) *txWaiter {
	return &txWaiter{
		tx:     tx,
		waitCh: make(chan struct{}),
		timeCh: time.After(timeout),
	}
}
