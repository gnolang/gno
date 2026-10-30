package gnoweb

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPkg = "gno.land/r/gnoland/wugnot"

// callTx is a transaction of one message per msgs entry, each a "Type
// pkgPath Func" triple.
func callTx(gas int, ok bool, msgs ...[3]string) indexer.Tx {
	tx := indexer.Tx{Success: ok, GasUsed: gas, Messages: make([]indexer.Message, len(msgs))}
	for i, m := range msgs {
		tx.Messages[i].Value.Type, tx.Messages[i].Value.PkgPath, tx.Messages[i].Value.Func = m[0], m[1], m[2]
	}
	return tx
}

func call(fn string) [3]string { return [3]string{"MsgCall", testPkg, fn} }

// Only successful single-call transactions into the package, with a gas
// use and a Go identifier for a function, are measured, at most gasUseCalls
// per function, newest first.
func TestGasUses(t *testing.T) {
	t.Parallel()

	txs := []indexer.Tx{
		callTx(9_000_000, false, call("Deposit")),                               // failed
		callTx(9_000_001, true, call("Deposit"), call("Withdraw")),              // two messages
		callTx(9_000_002, true, [3]string{"MsgCall", testPkg + "x", "Deposit"}), // other package
		callTx(9_000_003, true, [3]string{"MsgRun", testPkg, "Deposit"}),        // not a call
		callTx(9_000_004, true, [3]string{"UnexpectedMessage", "", ""}),         // what the filter let through (#6411)
		callTx(9_000_005, true), // no message
		callTx(9_000_006, true, call("")),
		callTx(9_000_007, true, call("Deposit<script>")),
		callTx(9_000_008, true, call("1Deposit")),
		callTx(9_000_009, true, call(strings.Repeat("D", maxGasUseFunc+1))),
		callTx(0, true, call("Deposit")),
		callTx(-1, true, call("Deposit")),
		callTx(1_400_000, true, call("Deposit")),
		callTx(1_300_000, true, call("Withdraw")),
		callTx(1_380_000, true, call("Deposit")),
		callTx(1_000, true, call(strings.Repeat("D", maxGasUseFunc))),
	}
	for i := range gasUseCalls {
		txs = append(txs, callTx(1_000_000+i, true, call("Transfer")))
	}
	txs = append(txs, callTx(5_000_000, true, call("Transfer"))) // the 21st: too old

	require.Equal(t, map[string]components.GasUse{
		"Deposit":                          {Max: 1_400_000, Calls: 2},
		"Withdraw":                         {Max: 1_300_000, Calls: 1},
		"Transfer":                         {Max: 1_000_000 + gasUseCalls - 1, Calls: gasUseCalls},
		strings.Repeat("D", maxGasUseFunc): {Max: 1_000, Calls: 1},
	}, gasUses(txs, testPkg))
}

// stubRecent answers RecentByPackage with txs or err, counting calls; when
// gate is set, each call waits on it.
type stubRecent struct {
	mu    sync.Mutex
	txs   []indexer.Tx
	err   error
	gate  chan struct{}
	calls int
}

func (s *stubRecent) set(txs []indexer.Tx, err error, gate chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.txs, s.err, s.gate = txs, err, gate
}

func (s *stubRecent) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *stubRecent) recent(_ context.Context, pkgPath string, limit int) ([]indexer.Tx, error) {
	s.mu.Lock()
	s.calls++
	txs, err, gate := s.txs, s.err, s.gate
	s.mu.Unlock()
	if gate != nil {
		<-gate
	}
	if limit != gasUseTxs {
		return nil, errors.New("limit " + strconv.Itoa(limit))
	}
	return txs, err
}

func deposits(gas ...int) []indexer.Tx {
	var txs []indexer.Tx
	for _, g := range gas {
		txs = append(txs, callTx(g, true, call("Deposit")))
	}
	return txs
}

// fakeClock is a gasUseCache clock the test moves.
type fakeClock struct{ ns atomic.Int64 }

func (f *fakeClock) now() time.Time      { return time.Unix(0, f.ns.Load()) }
func (f *fakeClock) add(d time.Duration) { f.ns.Add(int64(d)) }
func newGasUseCache(idx *stubRecent, clock *fakeClock) *gasUseCache {
	return &gasUseCache{recent: idx.recent, logger: newDiscardLogger(), now: clock.now}
}

// settle waits until pkgPath's entry is fresh.
func (c *gasUseCache) settle(t *testing.T, pkgPath string) {
	t.Helper()
	require.Eventually(t, func() bool {
		_, ok := c.lookup(pkgPath)
		return ok
	}, 5*time.Second, time.Millisecond)
}

// A page never waits: the first gets nothing while the read runs, the next
// get its answer, reused within gasUseTTL; past it, callers get the stale
// answer at once while one read refreshes it. A partial read is used.
func TestGasUseCache(t *testing.T) {
	t.Parallel()

	gate := make(chan struct{})
	idx := &stubRecent{txs: deposits(1_400_000, 1_380_000, 1_360_000), gate: gate}
	clock := &fakeClock{}
	c := newGasUseCache(idx, clock)
	ctx := context.Background()
	first := components.GasUse{Max: 1_400_000, Calls: 3}

	require.Nil(t, c.get(ctx, testPkg), "first read in flight")
	close(gate)
	c.settle(t, testPkg)
	require.Equal(t, first, c.get(ctx, testPkg)["Deposit"])
	clock.add(gasUseTTL - time.Second)
	require.Equal(t, first, c.get(ctx, testPkg)["Deposit"])
	require.Equal(t, 1, idx.count(), "reads within the TTL")

	// Stale: every caller gets the old answer at once while one read hangs.
	gate = make(chan struct{})
	idx.set(deposits(2_000_000, 2_000_000, 2_000_000), nil, gate)
	clock.add(time.Second)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.Equal(t, first, c.get(ctx, testPkg)["Deposit"], "during a refresh")
		}()
	}
	wg.Wait()
	close(gate)
	c.settle(t, testPkg)
	require.Equal(t, 2, idx.count(), "one refresh for all stale callers")
	require.Equal(t, int64(2_000_000), c.get(ctx, testPkg)["Deposit"].Max)

	idx.set(deposits(1_000_000), indexer.ErrPartial, nil)
	clock.add(gasUseTTL)
	c.get(ctx, testPkg)
	c.settle(t, testPkg)
	require.Equal(t, components.GasUse{Max: 1_000_000, Calls: 1}, c.get(ctx, testPkg)["Deposit"])
}

// A failed read is kept, as nothing, for gasUseRetry, then read again.
func TestGasUseCacheRetry(t *testing.T) {
	t.Parallel()

	idx := &stubRecent{err: errors.New("indexer down")}
	clock := &fakeClock{}
	c := newGasUseCache(idx, clock)
	ctx := context.Background()

	c.get(ctx, testPkg)
	c.settle(t, testPkg)
	clock.add(gasUseRetry - time.Second)
	require.Nil(t, c.get(ctx, testPkg))
	require.Equal(t, 1, idx.count())

	idx.set(deposits(1_400_000, 1_400_000, 1_400_000), nil, nil)
	clock.add(time.Second)
	c.get(ctx, testPkg)
	c.settle(t, testPkg)
	require.Equal(t, 2, idx.count())
	require.Equal(t, components.GasUse{Max: 1_400_000, Calls: 3}, c.get(ctx, testPkg)["Deposit"])
}

// At most maxGasUseReads run at once; a page past the cap skips its read
// without marking the package, so a later page reads it.
func TestGasUseCacheReadCap(t *testing.T) {
	t.Parallel()

	gate := make(chan struct{})
	idx := &stubRecent{gate: gate}
	c := newGasUseCache(idx, &fakeClock{})
	ctx := context.Background()

	for i := range maxGasUseReads {
		c.get(ctx, testPkg+strconv.Itoa(i))
	}
	require.Eventually(t, func() bool { return idx.count() == maxGasUseReads }, 5*time.Second, time.Millisecond)
	extra := testPkg + "x"
	c.get(ctx, extra)
	<-c.sf.DoChan(extra, func() (any, error) { return nil, nil }) // joins the skipped read until it returns
	close(gate)
	for i := range maxGasUseReads {
		c.settle(t, testPkg+strconv.Itoa(i))
	}
	_, ok := c.lookup(extra)
	require.False(t, ok, "skipped package marked")
	require.Equal(t, maxGasUseReads, idx.count())

	c.get(ctx, extra)
	c.settle(t, extra)
	require.Equal(t, maxGasUseReads+1, idx.count())
}

// Without an indexer nothing is read; with one, a full cache is cleared for
// a new package, not for a refresh.
func TestGasUseCacheBounds(t *testing.T) {
	t.Parallel()

	require.Nil(t, (&gasUseCache{}).get(context.Background(), testPkg), "no indexer")

	clock := &fakeClock{}
	c := newGasUseCache(&stubRecent{}, clock)
	c.entries = make(map[string]gasUseEntry)
	for i := range maxGasUsePkgs - 1 {
		c.entries[testPkg+strconv.Itoa(i)] = gasUseEntry{expires: clock.now().Add(time.Hour)}
	}
	c.get(context.Background(), testPkg)
	c.settle(t, testPkg)
	require.Len(t, c.entries, maxGasUsePkgs, "full, not over")

	clock.add(gasUseTTL)
	c.get(context.Background(), testPkg)
	c.settle(t, testPkg)
	require.Len(t, c.entries, maxGasUsePkgs, "a refresh clears nothing")

	c.get(context.Background(), testPkg+"x")
	c.settle(t, testPkg+"x")
	c.mu.Lock()
	defer c.mu.Unlock()
	require.Len(t, c.entries, 1, "cleared for a new package")
}
