package core

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTxDispatcher_ClosedSubscriptionDoesNotPanic regression-tests the case
// where the event switch closes the txDispatcher's subscription channel on
// shutdown (via events.SubscribeFilteredOn's <-evsw.Quit() branch). The
// listenRoutine used to panic with "txDispatcher subscription unexpectedly
// closed"; it should now exit cleanly.
func TestTxDispatcher_ClosedSubscriptionDoesNotPanic(t *testing.T) {
	t.Parallel()

	sub := make(chan events.Event)
	td := &txDispatcher{
		evsw:       events.NewEventSwitch(),
		listenerID: "test",
		sub:        sub,
		waiters:    make(map[string][]*txWaiter),
	}
	td.BaseService = *service.NewBaseService(nil, "txDispatcher", td)
	require.NoError(t, td.Start())

	done := make(chan struct{})
	go func() {
		td.Wait()
		close(done)
	}()

	close(sub)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("listenRoutine did not exit after subscription closed")
	}
}

// TestTxDispatcher_EventSwitchShutdown exercises the exact scenario reported
// in CI: FireEvent continues to invoke listener callbacks after evsw.Stop(),
// and when the unbuffered send would block while evsw.Quit() is closed, the
// subscribe callback closes the subscription channel. The listenRoutine must
// not panic.
func TestTxDispatcher_EventSwitchShutdown(t *testing.T) {
	t.Parallel()

	evsw := events.NewEventSwitch()
	require.NoError(t, evsw.Start())

	td := newTxDispatcher(evsw)
	t.Cleanup(func() {
		if td.IsRunning() {
			td.Stop()
		}
	})

	// We need to drive the event switch into a state where its listener
	// callback must take the <-evsw.Quit() branch and close the sub
	// channel. That requires: (a) listenRoutine not reading from ch when
	// the callback runs, and (b) evsw.Quit() already closed by then.
	//
	// Sequence:
	//  1. Lock td.mtx.
	//  2. Fire event #1. listenRoutine receives it, then blocks on
	//     notifyTxEvent's td.mtx.Lock().
	//  3. Fire event #2. listenRoutine is blocked, so its callback's
	//     `ch <- event` send cannot proceed — the callback waits in its
	//     own select.
	//  4. Stop the evsw. evsw.Quit() closes, waking event #2's callback,
	//     which now picks <-evsw.Quit() and calls close(ch).
	//  5. Unlock td.mtx. listenRoutine returns to its select, receives
	//     from closed ch, sees ok=false. Pre-fix: panic. Post-fix:
	//     return cleanly and Stop the dispatcher.
	tx := types.Tx("stall-tx")

	td.mtx.Lock()

	firstDone := make(chan struct{})
	go func() {
		evsw.FireEvent(types.EventTx{Result: types.TxResult{Tx: tx}})
		close(firstDone)
	}()
	// event #1's callback returns as soon as listenRoutine receives.
	select {
	case <-firstDone:
	case <-time.After(2 * time.Second):
		t.Fatal("first FireEvent never returned — listenRoutine not receiving")
	}
	// At this point listenRoutine is inside notifyTxEvent, blocking on
	// td.mtx (owned by the test).

	secondDone := make(chan struct{})
	go func() {
		defer close(secondDone)
		defer func() { _ = recover() }() // tolerate preexisting SubscribeFilteredOn panics
		evsw.FireEvent(types.EventTx{Result: types.TxResult{Tx: tx}})
	}()
	// Give event #2's callback time to enter its select and start
	// blocking on the send (listenRoutine isn't receiving).
	time.Sleep(50 * time.Millisecond)

	// Stop the evsw now — this wakes event #2's callback onto the Quit
	// branch, which calls close(ch).
	evsw.Stop()

	select {
	case <-secondDone:
	case <-time.After(2 * time.Second):
		t.Fatal("second FireEvent never returned after evsw.Stop")
	}

	// Release the mutex. listenRoutine finishes event #1 and observes
	// the closed subscription.
	td.mtx.Unlock()

	select {
	case <-td.Quit():
		// Expected: listenRoutine stopped the dispatcher cleanly.
	case <-time.After(2 * time.Second):
		t.Fatal("listenRoutine did not exit after subscription closed")
	}
}

// newTestTxDispatcher returns a dispatcher and its event switch, both stopped
// when the test ends.
func newTestTxDispatcher(t *testing.T) *txDispatcher {
	t.Helper()

	evsw := events.NewEventSwitch()
	require.NoError(t, evsw.Start())
	t.Cleanup(func() { evsw.Stop() })

	td := newTxDispatcher(evsw)
	t.Cleanup(func() {
		if td.IsRunning() {
			td.Stop()
		}
	})
	return td
}

// TestTxDispatcher_DuplicateTxEvent covers a block carrying one tx at two
// indexes: the second EventTx used to close an already-closed waitCh.
func TestTxDispatcher_DuplicateTxEvent(t *testing.T) {
	t.Parallel()

	td := newTestTxDispatcher(t)

	// require off the test goroutine would call t.FailNow there, so the result
	// comes back over a channel.
	tx := types.Tx("duplicate-me")
	type outcome struct {
		res types.TxResult
		err error
	}
	result := make(chan outcome, 1)
	go func() {
		res, err := td.getTxResult(tx, 30*time.Second, nil)
		result <- outcome{res, err}
	}()
	require.Eventually(t, func() bool {
		td.mtx.Lock()
		defer td.mtx.Unlock()
		return len(td.waiters[string(tx)]) == 1
	}, 2*time.Second, 5*time.Millisecond, "waiter never registered")

	event := func(index uint32) types.EventTx {
		return types.EventTx{Result: types.TxResult{Height: 9, Index: index, Tx: tx}}
	}
	td.notifyTxEvent(event(0))

	select {
	case got := <-result:
		require.NoError(t, got.err)
		require.EqualValues(t, 0, got.res.Index)
	case <-time.After(2 * time.Second):
		t.Fatal("waiter never resolved")
	}

	td.mtx.Lock()
	stillWaiting := len(td.waiters[string(tx)]) > 0
	td.mtx.Unlock()
	assert.False(t, stillWaiting, "resolved waiter retains the tx bytes")

	// The second copy of the same tx must be a no-op.
	require.NotPanics(t, func() { td.notifyTxEvent(event(1)) })
}

// TestTxDispatcher_TimeoutForgetsWaiter pins getTxResult's docstring: a node
// serving broadcast_tx_commit keeps no bytes of a tx that timed out.
func TestTxDispatcher_TimeoutForgetsWaiter(t *testing.T) {
	t.Parallel()

	td := newTestTxDispatcher(t)

	tx := types.Tx("never-committed")
	_, err := td.getTxResult(tx, 10*time.Millisecond, nil)
	require.EqualError(t, err, "request timeout")

	td.mtx.Lock()
	stillWaiting := len(td.waiters[string(tx)]) > 0
	td.mtx.Unlock()
	require.False(t, stillWaiting, "timed-out waiter retains the tx bytes")
}

// TestTxDispatcher_SecondCallerSurvivesFirstTimeout pins that one caller's
// timeout leaves another waiting on the same tx reachable from notifyTxEvent.
func TestTxDispatcher_SecondCallerSurvivesFirstTimeout(t *testing.T) {
	t.Parallel()

	td := newTestTxDispatcher(t)

	tx := types.Tx("two-callers")
	type outcome struct {
		res types.TxResult
		err error
	}

	// The short caller registers first and times out almost at once.
	short := make(chan outcome, 1)
	go func() {
		res, err := td.getTxResult(tx, 20*time.Millisecond, nil)
		short <- outcome{res, err}
	}()
	require.Eventually(t, func() bool { return waiterCount(td, tx) == 1 },
		2*time.Second, time.Millisecond, "short caller never registered")

	// The patient caller joins while the short one is still waiting.
	patient := make(chan outcome, 1)
	go func() {
		res, err := td.getTxResult(tx, 30*time.Second, nil)
		patient <- outcome{res, err}
	}()
	require.Eventually(t, func() bool { return waiterCount(td, tx) == 2 },
		2*time.Second, time.Millisecond, "patient caller never registered")

	select {
	case got := <-short:
		require.EqualError(t, got.err, "request timeout")
	case <-time.After(2 * time.Second):
		t.Fatal("short caller never timed out")
	}

	// The tx commits after that timeout. The patient caller still gets it.
	td.notifyTxEvent(types.EventTx{Result: types.TxResult{Height: 9, Tx: tx}})
	select {
	case got := <-patient:
		require.NoError(t, got.err)
		require.EqualValues(t, 9, got.res.Height)
	case <-time.After(2 * time.Second):
		t.Fatal("patient caller stranded by the other caller's timeout")
	}
}

// waiterCount reports how many callers are registered for tx.
func waiterCount(td *txDispatcher, tx types.Tx) int {
	td.mtx.Lock()
	defer td.mtx.Unlock()

	return len(td.waiters[string(tx)])
}
