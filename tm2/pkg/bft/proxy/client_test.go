package proxy

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
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

	// The bound is GOMAXPROCS, so ask for exactly that many rather than a
	// number that happens to fit on this machine. Reading the package var is
	// safe; writing it would race every other test that builds a client.
	callers := maxConcurrentQueries
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

	// The limiter is passed explicitly rather than by setting
	// maxConcurrentQueries, which is a package-level var that every other test
	// reads while building a client. This is the same construction
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
// produce a zero-capacity channel, which would deadlock on first Lock. There is
// no unbounded mode: a bound that cannot be expressed defaults to one slot.
func TestQueryLimiterNeverZeroCapacity(t *testing.T) {
	t.Parallel()

	for _, n := range []int{0, -1} {
		l := newQueryLimiter(n)
		l.Lock()
		l.Unlock()
	}
}
