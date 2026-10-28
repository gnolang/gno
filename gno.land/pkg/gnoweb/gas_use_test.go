package gnoweb

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
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

// Only successful single-call transactions into the package are measured,
// at most gasUseCalls per function, newest first.
func TestGasUses(t *testing.T) {
	t.Parallel()

	txs := []indexer.Tx{
		callTx(9_000_000, false, call("Deposit")),                               // failed
		callTx(9_000_001, true, call("Deposit"), call("Withdraw")),              // two messages
		callTx(9_000_002, true, [3]string{"MsgCall", testPkg + "x", "Deposit"}), // other package
		callTx(9_000_003, true, [3]string{"MsgRun", testPkg, "Deposit"}),        // not a call
		callTx(9_000_004, true, [3]string{"UnexpectedMessage", "", ""}),         // what the filter let through (#6411)
		callTx(9_000_005, true), // no message
		callTx(1_400_000, true, call("Deposit")),
		callTx(1_300_000, true, call("Withdraw")),
		callTx(1_380_000, true, call("Deposit")),
	}
	for i := range gasUseCalls {
		txs = append(txs, callTx(1_000_000+i, true, call("Transfer")))
	}
	txs = append(txs, callTx(5_000_000, true, call("Transfer"))) // the 21st: too old

	got := gasUses(txs, testPkg)
	want := map[string]components.GasUse{
		"Deposit":  {Max: 1_400_000, Calls: 2},
		"Withdraw": {Max: 1_300_000, Calls: 1},
		"Transfer": {Max: 1_000_000 + gasUseCalls - 1, Calls: gasUseCalls},
	}
	if len(got) != len(want) {
		t.Fatalf("gasUses = %v, want %v", got, want)
	}
	for fn, w := range want {
		if got[fn] != w {
			t.Errorf("%s: %+v, want %+v", fn, got[fn], w)
		}
	}
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

// expire ages pkgPath's entry past gasUseTTL.
func (c *gasUseCache) expire(pkgPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[pkgPath].at = c.entries[pkgPath].at.Add(-gasUseTTL)
}

// settle waits for pkgPath's read to finish.
func (c *gasUseCache) settle(t *testing.T, pkgPath string) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(time.Millisecond) {
		c.mu.Lock()
		busy := c.entries[pkgPath].busy
		c.mu.Unlock()
		if !busy {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("read did not finish")
		}
	}
}

// The first read is waited for; within gasUseTTL the answer is reused; past
// it, callers get the stale answer at once while one read refreshes it. A
// failed read keeps the last answer, and a partial one is used.
func TestGasUseCache(t *testing.T) {
	t.Parallel()

	idx := &stubRecent{txs: deposits(1_400_000, 1_380_000, 1_360_000)}
	c := &gasUseCache{recent: idx.recent, logger: newDiscardLogger()}
	ctx := context.Background()
	first := components.GasUse{Max: 1_400_000, Calls: 3}

	if got := c.get(ctx, testPkg)["Deposit"]; got != first {
		t.Fatalf("first read: %+v, want %+v", got, first)
	}
	if got := c.get(ctx, testPkg)["Deposit"]; got != first || idx.count() != 1 {
		t.Fatalf("within the TTL: %+v after %d reads, want %+v after 1", got, idx.count(), first)
	}

	// Stale: every caller gets the old answer at once while one read hangs.
	gate := make(chan struct{})
	idx.set(deposits(2_000_000, 2_000_000, 2_000_000), nil, gate)
	c.expire(testPkg)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := c.get(ctx, testPkg)["Deposit"]; got != first {
				t.Errorf("during a refresh: %+v, want the stale %+v", got, first)
			}
		}()
	}
	wg.Wait()
	close(gate)
	c.settle(t, testPkg)
	if n := idx.count(); n != 2 {
		t.Fatalf("%d reads, want 2: one refresh for all stale callers", n)
	}
	if got := c.get(ctx, testPkg)["Deposit"]; got.Max != 2_000_000 {
		t.Fatalf("after the refresh: %+v, want 2 M", got)
	}

	// A failed read keeps the last answer; a partial one replaces it.
	idx.set(nil, errors.New("indexer down"), nil)
	c.expire(testPkg)
	c.get(ctx, testPkg)
	c.settle(t, testPkg)
	if got := c.get(ctx, testPkg)["Deposit"]; got.Max != 2_000_000 || idx.count() != 3 {
		t.Fatalf("after a failure: %+v after %d reads, want the last answer after 3", got, idx.count())
	}
	idx.set(deposits(1_000_000), indexer.ErrPartial, nil)
	c.expire(testPkg)
	c.get(ctx, testPkg)
	c.settle(t, testPkg)
	if got := c.get(ctx, testPkg)["Deposit"]; got != (components.GasUse{Max: 1_000_000, Calls: 1}) {
		t.Fatalf("after a partial read: %+v, want its one call", got)
	}
}

// A reader whose request is gone does not wait for a slow first read, which
// still fills the cache for the next one.
func TestGasUseCacheSlowFirstRead(t *testing.T) {
	t.Parallel()

	gate := make(chan struct{})
	idx := &stubRecent{txs: deposits(1_400_000, 1_400_000, 1_400_000), gate: gate}
	c := &gasUseCache{recent: idx.recent, logger: newDiscardLogger()}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := c.get(ctx, testPkg); got != nil {
		t.Fatalf("cancelled reader got %v, want nothing yet", got)
	}
	close(gate)
	c.settle(t, testPkg)
	if got := c.get(context.Background(), testPkg)["Deposit"]; got.Calls != 3 || idx.count() != 1 {
		t.Fatalf("next reader: %+v after %d reads, want 3 calls after 1", got, idx.count())
	}
}

// Without an indexer nothing is read; with one, the cache holds at most
// maxGasUsePkgs packages.
func TestGasUseCacheBounds(t *testing.T) {
	t.Parallel()

	if got := (&gasUseCache{}).get(context.Background(), testPkg); got != nil {
		t.Fatalf("no indexer: %v, want nil", got)
	}

	idx := &stubRecent{}
	c := &gasUseCache{recent: idx.recent, logger: newDiscardLogger()}
	for i := range maxGasUsePkgs + 10 {
		c.get(context.Background(), testPkg+strconv.Itoa(i))
	}
	c.mu.Lock()
	n := len(c.entries)
	c.mu.Unlock()
	if n != maxGasUsePkgs {
		t.Fatalf("%d packages cached, want %d", n, maxGasUsePkgs)
	}
}
