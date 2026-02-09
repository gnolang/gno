package mempool

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/service"
)

// This code was moved over from the old Tendermint RPC implementation, and slightly cleaned up.
// If time allows, we should remove it altogether, and figure out a better mechanism for transaction waiting
type txDispatcher struct {
	service.BaseService

	evsw       events.EventSwitch
	listenerID string
	sub        <-chan events.Event

	timeout time.Duration

	mtx     sync.Mutex
	waiters map[string]*txWaiter // string(tx) -> waiter shared by all callers
}

func newTxDispatcher(evsw events.EventSwitch, timeout time.Duration) *txDispatcher {
	listenerID := fmt.Sprintf("txDispatcher#%v", random.RandStr(6))
	sub := events.SubscribeToEvent(evsw, listenerID, types.EventTx{})

	td := &txDispatcher{
		evsw:       evsw,
		listenerID: listenerID,
		sub:        sub,
		timeout:    timeout,
		waiters:    make(map[string]*txWaiter),
	}

	td.BaseService = *service.NewBaseService(nil, "txDispatcher", td)

	if err := td.Start(); err != nil {
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
	key := string(txEvent.Result.Tx)

	td.mtx.Lock()
	waiter, ok := td.waiters[key]
	if !ok {
		td.mtx.Unlock()
		return
	}

	delete(td.waiters, key)

	waiter.res = txEvent.Result
	close(waiter.done)

	td.mtx.Unlock()
}

// getTxResult blocks until:
//   - the tx result arrives from events, OR
//   - the dispatcher timeout expires, OR
//   - the caller's quit channel fires (if non-nil).
//
// All callers waiting on the same tx share the same waiter and get the same result
func (td *txDispatcher) getTxResult(tx types.Tx, quit chan struct{}) (types.TxResult, error) {
	key := string(tx)

	td.mtx.Lock()
	waiter, ok := td.waiters[key]
	if !ok {
		waiter = newTxWaiter()
		td.waiters[key] = waiter
	}
	td.mtx.Unlock()

	timeout := time.After(td.timeout)

	select {
	case <-waiter.done:
		return waiter.res, nil

	case <-timeout:
		td.mtx.Lock()
		delete(td.waiters, key)
		td.mtx.Unlock()

		return types.TxResult{}, errors.New("request timeout")

	case <-quit:
		return types.TxResult{}, errors.New("caller quit")
	}
}

type txWaiter struct {
	done chan struct{}
	res  types.TxResult
}

func newTxWaiter() *txWaiter {
	return &txWaiter{
		done: make(chan struct{}),
	}
}
