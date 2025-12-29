package core

import (
	"fmt"
	"sync"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// -----------------------------------------------------------------------------
// NOTE: tx should be signed, but this is only checked at the app level (not by Tendermint!)

// Returns right away, with no response. Does not wait for CheckTx nor
// DeliverTx results.
//
// If you want to be sure that the transaction is included in a block, you can
// subscribe for the result using JSONRPC via a websocket. See
// https://docs.tendermint.com/v0.34/tendermint-core/subscription.html
// If you haven't received anything after a couple of blocks, resend it. If the
// same happens again, send it to some other node. A few reasons why it could
// happen:
//
// 1. malicious node can drop or pretend it had committed your tx
// 2. malicious proposer (not necessary the one you're communicating with) can
// drop transactions, which might become valid in the future
// (https://github.com/tendermint/tendermint/issues/3322)
// 3. node can be offline
//
// Please refer to
// https://docs.tendermint.com/v0.34/tendermint-core/using-tendermint.html#formatting
// for formatting/encoding rules.
//
// ```shell
// curl 'localhost:26657/broadcast_tx_async?tx="123"'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:26657", "/websocket")
// err := client.Start()
//
//	if err != nil {
//	  // handle error
//	}
//
// defer client.Stop()
// result, err := client.BroadcastTxAsync("123")
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
//
//	{
//		"error": "",
//		"result": {
//			"hash": "E39AAB7A537ABAA237831742DCE1117F187C3C52",
//			"log": "",
//			"data": "",
//			"code": "0"
//		},
//		"id": "",
//		"jsonrpc": "2.0"
//	}
//
// ```
//
// ### Query Parameters
//
// | Parameter | Type | Default | Required | Description     |
// |-----------+------+---------+----------+-----------------|
// | tx        | Tx   | nil     | true     | The transaction |
func BroadcastTxAsync(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "BroadcastTxAsync")
	defer span.End()
	err := mempool.CheckTx(tx, nil)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultBroadcastTx{Hash: tx.Hash()}, nil
}

// Returns with the response from CheckTx. Does not wait for DeliverTx result.
//
// If you want to be sure that the transaction is included in a block, you can
// subscribe for the result using JSONRPC via a websocket. See
// https://docs.tendermint.com/v0.34/tendermint-core/subscription.html
// If you haven't received anything after a couple of blocks, resend it. If the
// same happens again, send it to some other node. A few reasons why it could
// happen:
//
// 1. malicious node can drop or pretend it had committed your tx
// 2. malicious proposer (not necessary the one you're communicating with) can
// drop transactions, which might become valid in the future
// (https://github.com/tendermint/tendermint/issues/3322)
//
// Please refer to
// https://docs.tendermint.com/v0.34/tendermint-core/using-tendermint.html#formatting
// for formatting/encoding rules.
//
// ```shell
// curl 'localhost:26657/broadcast_tx_sync?tx="456"'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:26657", "/websocket")
// err := client.Start()
//
//	if err != nil {
//	  // handle error
//	}
//
// defer client.Stop()
// result, err := client.BroadcastTxSync("456")
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
//
//	{
//		"jsonrpc": "2.0",
//		"id": "",
//		"result": {
//			"code": "0",
//			"data": "",
//			"log": "",
//			"hash": "0D33F2F03A5234F38706E43004489E061AC40A2E"
//		},
//		"error": ""
//	}
//
// ```
//
// ### Query Parameters
//
// | Parameter | Type | Default | Required | Description     |
// |-----------+------+---------+----------+-----------------|
// | tx        | Tx   | nil     | true     | The transaction |
func BroadcastTxSync(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "BroadcastTxSync")
	defer span.End()
	resCh := make(chan abci.Response, 1)
	err := mempool.CheckTx(tx, func(res abci.Response) {
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

// Returns with the responses from CheckTx and DeliverTx.
//
// IMPORTANT: use only for testing and development. In production, use
// BroadcastTxSync or BroadcastTxAsync. You can subscribe for the transaction
// result using JSONRPC via a websocket. See
// https://docs.tendermint.com/v0.34/tendermint-core/subscription.html
//
// CONTRACT: only returns error if mempool.CheckTx() errs or if we timeout
// waiting for tx to commit.
//
// If CheckTx or DeliverTx fail, no error will be returned, but the returned result
// will contain a non-OK ABCI code.
//
// Please refer to
// https://docs.tendermint.com/v0.34/tendermint-core/using-tendermint.html#formatting
// for formatting/encoding rules.
//
// ```shell
// curl 'localhost:26657/broadcast_tx_commit?tx="789"'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:26657", "/websocket")
// err := client.Start()
//
//	if err != nil {
//	  // handle error
//	}
//
// defer client.Stop()
// result, err := client.BroadcastTxCommit("789")
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
//
//	{
//		"error": "",
//		"result": {
//			"height": "26682",
//			"hash": "75CA0F856A4DA078FC4911580360E70CEFB2EBEE",
//			"deliver_tx": {
//				"log": "",
//				"data": "",
//				"code": "0"
//			},
//			"check_tx": {
//				"log": "",
//				"data": "",
//				"code": "0"
//			}
//		},
//		"id": "",
//		"jsonrpc": "2.0"
//	}
//
// ```
//
// ### Query Parameters
//
// | Parameter | Type | Default | Required | Description     |
// |-----------+------+---------+----------+-----------------|
// | tx        | Tx   | nil     | true     | The transaction |
func BroadcastTxCommit(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "BroadcastTxCommit")
	defer span.End()
	// Broadcast tx and wait for CheckTx result
	checkTxResCh := make(chan abci.Response, 1)
	err := mempool.CheckTx(tx, func(res abci.Response) {
		checkTxResCh <- res
	})
	if err != nil {
		logger.Error("Error on broadcastTxCommit", "err", err)
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
	txRes, err := gTxDispatcher.getTxResult(tx, nil)
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

// Get unconfirmed transactions (maximum ?limit entries) including their number.
//
// ```shell
// curl 'localhost:26657/unconfirmed_txs'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:26657", "/websocket")
// err := client.Start()
//
//	if err != nil {
//	  // handle error
//	}
//
// defer client.Stop()
// result, err := client.UnconfirmedTxs()
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
//
//	{
//	  "result" : {
//	      "txs" : [],
//	      "total_bytes" : "0",
//	      "n_txs" : "0",
//	      "total" : "0"
//	    },
//	    "jsonrpc" : "2.0",
//	    "id" : ""
//	  }
//
// ```
//
// ### Query Parameters
//
// | Parameter | Type | Default | Required | Description                          |
// |-----------+------+---------+----------+--------------------------------------|
// | limit     | int  | 30      | false    | Maximum number of entries (max: 100) |
// ```
func UnconfirmedTxs(ctx *rpctypes.Context, limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "UnconfirmedTxs")
	defer span.End()
	// reuse per_page validator
	limit = validatePerPage(limit)

	txs := mempool.ReapMaxTxs(limit)
	return &ctypes.ResultUnconfirmedTxs{
		Count:      len(txs),
		Total:      mempool.Size(),
		TotalBytes: mempool.TxsBytes(),
		Txs:        txs,
	}, nil
}

// Get number of unconfirmed transactions.
//
// ```shell
// curl 'localhost:26657/num_unconfirmed_txs'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:26657", "/websocket")
// err := client.Start()
// if err != nil {
// // handle error
// }
// defer client.Stop()
// result, err := client.UnconfirmedTxs()
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
//
//	{
//	  "jsonrpc" : "2.0",
//	  "id" : "",
//	  "result" : {
//	    "n_txs" : "0",
//	    "total_bytes" : "0",
//	    "total" : "0"
//	    "txs" : null,
//	  }
//	}
//
// ```
func NumUnconfirmedTxs(ctx *rpctypes.Context) (*ctypes.ResultUnconfirmedTxs, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "NumUnconfirmedTxs")
	defer span.End()
	return &ctypes.ResultUnconfirmedTxs{
		Count:      mempool.Size(),
		Total:      mempool.Size(),
		TotalBytes: mempool.TxsBytes(),
	}, nil
}

// ----------------------------------------
// txListener

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
				td.Stop()
				panic("txDispatcher subscription unexpectedly closed")
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
	} else {
		waiter.txRes = txEvent.Result
		close(waiter.waitCh)
	}
}

// blocking
// If the tx is already being waited on, returns the result from the original request.
// Upon result or timeout, the tx is forgotten from txDispatcher, and can be re-requested.
// If the tx times out, an error is returned.
// Quit can optionally be provided to terminate early (e.g. if the caller disconnects).
func (td *txDispatcher) getTxResult(tx types.Tx, quit chan struct{}) (types.TxResult, error) {
	// Get or create waiter.
	td.mtx.Lock()
	waiter, ok := td.waiters[string(tx)]
	if !ok {
		waiter = newTxWaiter(tx)
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

func newTxWaiter(tx types.Tx) *txWaiter {
	return &txWaiter{
		tx:     tx,
		waitCh: make(chan struct{}),
		timeCh: time.After(config.TimeoutBroadcastTxCommit),
	}
}
