package proxy

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/meminfo"
)

// gateApp is an application whose Query blocks until released. It turns the
// connection's concurrency into something a test can assert structurally rather
// than statistically: if the connection admits fewer callers than the gate is
// waiting for, the gate never opens and the test fails on a timeout instead of
// merely being slower.
type gateApp struct {
	abci.BaseApplication

	arrived  chan struct{}
	release  chan struct{}
	inFlight atomic.Int64
	peak     atomic.Int64
}

func newGateApp() *gateApp {
	return &gateApp{
		arrived: make(chan struct{}, 1024),
		release: make(chan struct{}),
	}
}

func (g *gateApp) Query(abci.RequestQuery) abci.ResponseQuery {
	cur := g.inFlight.Add(1)
	for {
		old := g.peak.Load()
		if cur <= old || g.peak.CompareAndSwap(old, cur) {
			break
		}
	}
	defer g.inFlight.Add(-1)

	g.arrived <- struct{}{}
	<-g.release
	return abci.ResponseQuery{}
}

// waitArrivals waits for n callers to be simultaneously inside Query.
func (g *gateApp) waitArrivals(t *testing.T, n int, why string) {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for range n {
		select {
		case <-g.arrived:
		case <-deadline:
			t.Fatalf("%s: only %d of %d callers reached Application.Query; the connection is admitting fewer than expected",
				why, g.inFlight.Load(), n)
		}
	}
}

// fire issues n concurrent queries and returns a func that waits for them.
func fire(t *testing.T, client interface {
	QuerySync(abci.RequestQuery) (abci.ResponseQuery, error)
}, n int,
) func() {
	t.Helper()
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			_, _ = client.QuerySync(abci.RequestQuery{Path: "test"})
		})
	}
	return wg.Wait
}

// TestReadOnlyClientAdmitsConcurrentQueries is the structural counterpart to the
// gnoland integration stress test: it pins that the read-only connection admits
// more than one caller at a time, in a package with no database, no VM and no
// genesis. Reintroducing a per-call mutex here deadlocks this test rather than
// slowing it down.
func TestReadOnlyClientAdmitsConcurrentQueries(t *testing.T) {
	t.Parallel()

	// Ask for exactly the limit the connection starts at, rather than a
	// number that happens to fit on this machine.
	callers := min(runtime.GOMAXPROCS(0), queryLimiterFloor)
	if callers < 2 {
		t.Skip("GOMAXPROCS is 1: this machine cannot run two queries at once")
	}

	app := newGateApp()
	creator := NewLocalClientCreator(app)
	client, err := creator.NewReadOnlyABCIClient()
	require.NoError(t, err)
	require.NoError(t, client.Start())
	t.Cleanup(func() { _ = client.Stop() })

	wait := fire(t, client, callers)
	app.waitArrivals(t, callers, "read-only connection")
	close(app.release)
	wait()

	require.EqualValues(t, callers, app.peak.Load(),
		"expected %d simultaneous callers inside Application.Query", callers)
}

// TestReadOnlyClientRespectsConcurrencyBound pins the other half: the bound is
// real. With one slot the connection behaves exactly like the mutex it replaced.
func TestReadOnlyClientRespectsConcurrencyBound(t *testing.T) {
	t.Parallel()

	// The limiter is passed explicitly so the bound under test does not depend on
	// the machine's core count. This is the same construction
	// NewReadOnlyABCIClient performs.
	app := newGateApp()
	client := abcicli.NewLocalClient(newQueryLimiter(1), app)
	require.NoError(t, client.Start())
	t.Cleanup(func() { _ = client.Stop() })

	wait := fire(t, client, 4)

	// Exactly one caller may be inside Query. Wait for it, then confirm no
	// second one shows up while it is still held.
	app.waitArrivals(t, 1, "bounded connection")
	select {
	case <-app.arrived:
		t.Fatal("a second caller entered Application.Query while the bound was 1")
	case <-time.After(150 * time.Millisecond):
	}

	close(app.release)
	wait()

	require.EqualValues(t, 1, app.peak.Load(),
		"a one-slot limiter must serialise the connection")
}

// TestReadOnlyClientRejectsSetResponseCallback pins the one method the query
// connection's bound makes unsafe. localClient guards the write with the Locker
// it was given, which on this connection is a limiter admitting GOMAXPROCS
// callers, so the lock reads as synchronisation and provides none. The client
// has to say so rather than let the write through.
func TestReadOnlyClientRejectsSetResponseCallback(t *testing.T) {
	t.Parallel()

	creator := NewLocalClientCreator(abci.NewBaseApplication())
	client, err := creator.NewReadOnlyABCIClient()
	require.NoError(t, err)

	require.Panics(t, func() {
		client.SetResponseCallback(func(abci.Request, abci.Response) {})
	}, "the read-only connection accepted a response callback")

	// The mutating connections keep it: their Locker is a mutex, so the write
	// is excluded from completeRequest exactly as before.
	mutating, err := creator.NewABCIClient()
	require.NoError(t, err)
	require.NotPanics(t, func() {
		mutating.SetResponseCallback(func(abci.Request, abci.Response) {})
	}, "the consensus/mempool connection must still accept a response callback")
}

// TestMutatingClientSerialises pins that this change did not loosen the
// consensus and mempool connections: they still admit one caller at a time.
func TestMutatingClientSerialises(t *testing.T) {
	t.Parallel()

	app := newGateApp()
	creator := NewLocalClientCreator(app)
	client, err := creator.NewABCIClient()
	require.NoError(t, err)
	require.NoError(t, client.Start())
	t.Cleanup(func() { _ = client.Stop() })

	wait := fire(t, client, 4)

	app.waitArrivals(t, 1, "mutating connection")
	select {
	case <-app.arrived:
		t.Fatal("the mutating connection admitted a second concurrent caller")
	case <-time.After(150 * time.Millisecond):
	}

	close(app.release)
	wait()

	require.EqualValues(t, 1, app.peak.Load(),
		"consensus and mempool connections must stay serialised")
}

// TestConsensusAndQueryConnectionsDoNotContend pins the property PR5431 added
// and this change preserves: a query in flight must not block a consensus call.
// The query connection is gated open, and a consensus call has to complete
// while it is still parked inside the application.
func TestConsensusAndQueryConnectionsDoNotContend(t *testing.T) {
	t.Parallel()

	app := newGateApp()
	creator := NewLocalClientCreator(app)

	queryCli, err := creator.NewReadOnlyABCIClient()
	require.NoError(t, err)
	require.NoError(t, queryCli.Start())
	t.Cleanup(func() { _ = queryCli.Stop() })

	consCli, err := creator.NewABCIClient()
	require.NoError(t, err)
	require.NoError(t, consCli.Start())
	t.Cleanup(func() { _ = consCli.Stop() })

	wait := fire(t, queryCli, 1)
	app.waitArrivals(t, 1, "query connection")

	// Info does not go through gateApp.Query, so it returns immediately unless
	// it is blocked on a lock the parked query holds.
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = consCli.InfoSync(abci.RequestInfo{})
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("a consensus call blocked behind an in-flight query")
	}

	close(app.release)
	wait()
}

// TestQueryLimiterNeverZeroCapacity pins that a non-positive bound cannot
// produce a zero limit, which would deadlock on first Lock. There is
// no unbounded mode: a bound that cannot be expressed defaults to one slot.
func TestQueryLimiterNeverZeroCapacity(t *testing.T) {
	t.Parallel()

	for _, n := range []int{0, -1} {
		l := newQueryLimiter(n)
		l.Lock()
		l.Unlock()
	}
}

// TestQueryLimiterResize pins how the limit follows memory: up only while
// saturated with the reserve plus queryMemMargin free, down while the reserve
// is not, held in between, and never outside [min, max].
func TestQueryLimiterResize(t *testing.T) {
	t.Parallel()

	const gib = 1 << 30
	mem := meminfo.Info{Total: 16 * gib} // reserve 4 GiB; growing needs 5
	l := newQueryLimiter(2)
	l.max = 4
	l.readMem = func() (meminfo.Info, bool) { return mem, true }
	step := func(avail uint64, running int) int {
		mem.Available, l.running, l.lastCheck = avail, running, time.Time{}
		l.resize()
		return l.limit
	}

	require.Equal(t, 2, step(8*gib, 1), "a free slot is no reason to grow")
	require.Equal(t, 3, step(8*gib, 2), "saturated with memory to spare")
	require.Equal(t, 4, step(5*gib, 3), "the reserve plus the margin is enough")
	require.Equal(t, 4, step(8*gib, 4), "never past max")
	require.Equal(t, 4, step(4*gib, 4), "between the thresholds the limit holds")
	require.Equal(t, 3, step(3*gib, 4), "short of the reserve")
	require.Equal(t, 2, step(3*gib, 4))
	require.Equal(t, 2, step(0, 4), "never below min")

	l.limit = 3
	require.Equal(t, 3, step(5*gib-1, 3), "a byte short of the margin does not grow")

	// Readings are paced: a second one within the interval changes nothing.
	l.running = 3
	mem.Available = 8 * gib
	l.resize()
	require.Equal(t, 3, l.limit)
}

// TestQueryLimiterLockGrows pins that a caller arriving at a saturated limiter
// with memory to spare gets in, rather than waiting for a slot.
func TestQueryLimiterLockGrows(t *testing.T) {
	t.Parallel()

	l := newQueryLimiter(1)
	l.max = 2
	l.readMem = func() (meminfo.Info, bool) { return meminfo.Info{Total: 16 << 30, Available: 8 << 30}, true }

	l.Lock()
	l.lastCheck = time.Time{}
	done := make(chan struct{})
	go func() { l.Lock(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("second caller blocked although memory allowed a second slot")
	}
}

// TestQueryLimiterWaitersGrow pins that callers already queued count towards
// saturation: a burst that arrives at once still ramps as slots free up, though
// the reading after each Unlock lands in a waiter that finds a slot free.
func TestQueryLimiterWaitersGrow(t *testing.T) {
	t.Parallel()

	l := newQueryLimiter(1)
	l.max = 3
	l.readMem = func() (meminfo.Info, bool) { return meminfo.Info{Total: 16 << 30, Available: 8 << 30}, true }

	// Read while nothing runs, so the burst below meets a paced limiter.
	l.Lock()
	var admitted atomic.Int64
	for range 3 {
		go func() { l.Lock(); admitted.Add(1) }()
	}
	require.Eventually(t, func() bool {
		l.mu.Lock()
		defer l.mu.Unlock()
		return l.waiting == 3
	}, 5*time.Second, time.Millisecond)

	l.mu.Lock()
	l.lastCheck = time.Time{}
	l.mu.Unlock()
	// The woken waiter takes the freed slot, grows the limit to 2 for the two
	// still queued, and wakes one of them for it.
	l.Unlock()
	require.Eventually(t, func() bool { return admitted.Load() == 2 }, 5*time.Second, time.Millisecond,
		"the limit stayed at 1 with callers queued and memory to spare")
	l.mu.Lock()
	defer l.mu.Unlock()
	require.Equal(t, 2, l.limit)
}

// TestNewMemoryQueryLimiter pins the bounds the connection is built with: a
// ramp from the floor to GOMAXPROCS where memory can be read, and GOMAXPROCS
// fixed where it cannot.
func TestNewMemoryQueryLimiter(t *testing.T) {
	t.Parallel()

	most := runtime.GOMAXPROCS(0)

	unreadable := newMemoryQueryLimiter(func() (meminfo.Info, bool) { return meminfo.Info{}, false })
	require.Equal(t, most, unreadable.limit, "no reading: the cores are the bound")
	require.Equal(t, most, unreadable.min)
	require.Equal(t, most, unreadable.max)

	readable := newMemoryQueryLimiter(func() (meminfo.Info, bool) { return meminfo.Info{Total: 1, Available: 1}, true })
	floor := min(most, queryLimiterFloor)
	require.Equal(t, floor, readable.limit, "starts at the floor")
	require.Equal(t, floor, readable.min)
	require.Equal(t, most, readable.max)
	require.Equal(t, floor < most, readable.readMem != nil, "reads memory only when there is room to ramp")
}
