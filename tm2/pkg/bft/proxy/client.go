package proxy

import (
	"runtime"
	"sync"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/counter"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

var (
	// maxConcurrentQueries bounds how many calls may be inside the application
	// through the read-only connection at once.
	maxConcurrentQueries = max(runtime.GOMAXPROCS(0), 1)
)

// ClientCreator creates ABCI clients for the three Tendermint connections.
type ClientCreator interface {
	// NewABCIClient returns a client for mutating connections (consensus, mempool).
	// Calls on it are serialised against each other.
	NewABCIClient() (abcicli.Client, error)

	// NewReadOnlyABCIClient returns a client for the query connection. Calls on
	// it never block consensus, and they do not serialise against each other:
	// several may be inside the application at once.
	//
	// PRECONDITION on any implementation: the application must be goroutine-safe
	// for every method this connection reaches — QuerySync, InfoSync, EchoSync.
	// An implementation that cannot promise that must serialise the connection
	// itself; returning a client that admits concurrent callers into an
	// application that is not ready for them is a data race, not a slow query.
	NewReadOnlyABCIClient() (abcicli.Client, error)
}

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type localClientCreator struct {
	mtx sync.Mutex // shared by consensus and mempool connections
	app abci.Application
}

func NewLocalClientCreator(app abci.Application) ClientCreator {
	return &localClientCreator{app: app}
}

func (l *localClientCreator) NewABCIClient() (abcicli.Client, error) {
	return abcicli.NewLocalClient(&l.mtx, l.app), nil
}

// NewReadOnlyABCIClient returns a client whose calls do not contend with
// consensus or mempool, and do not serialise against each other: up to
// maxConcurrentQueries of them may be inside the application at once.
//
// This rests on an invariant: everything reachable from Application.Query and
// Application.Info is goroutine-safe (immutable DB snapshots, SyncGoMap
// cacheNodes sealed on publication, atomic last-block header, per-tx forked
// allocator). Breaking it reintroduces a data race.
//
// The returned client satisfies the full abcicli.Client interface, but only the
// three methods appconn.Query exposes — EchoSync, InfoSync, QuerySync — are safe
// to use on it. In particular SetResponseCallback is unsynchronised here: its
// write to the client's Callback field would race with the concurrent reads in
// completeRequest. Nothing calls it on this connection today — the appconn
// Query wrapper does not expose it — and nothing should.
func (l *localClientCreator) NewReadOnlyABCIClient() (abcicli.Client, error) {
	return abcicli.NewLocalClient(newQueryLimiter(maxConcurrentQueries), l.app), nil
}

// queryLimiter is the Locker handed to the read-only connection. It bounds how
// many calls may be inside the application at once instead of admitting one at
// a time.
//
// The bound is not about correctness — the invariant above is what makes
// concurrent queries safe — it is about work. Each query installs its own gas
// and allocation budget (maxGasQuery, maxAllocQuery in the VM keeper), so those
// cap a single query and never the sum. While the connection held a mutex the
// sum was the per-call cap, because one caller ran at a time; without a bound
// the only ceiling left is the RPC listener's MaxOpenConnections, which defaults
// to 900 and is nowhere near a machine's capacity to execute 900 simultaneous
// simulates. The limiter puts the ceiling back where it can be reasoned about.
//
// A limit of 1 makes this exactly the mutex it replaced.
type queryLimiter struct {
	slots chan struct{}
}

func newQueryLimiter(n int) sync.Locker {
	return &queryLimiter{slots: make(chan struct{}, max(n, 1))}
}

// Lock blocks until a slot is free. localClient defers every Unlock, so a slot
// is returned even when the application panics.
func (q *queryLimiter) Lock()   { q.slots <- struct{}{} }
func (q *queryLimiter) Unlock() { <-q.slots }

//-----------------------------------------------------------------
// DefaultClientCreator

// Returns the local application, or constructs a new one via proxy.
// This function is meant to work with config fields.
func DefaultClientCreator(local abci.Application, proxy string, transport, dbDir string) ClientCreator {
	if local != nil {
		// local applications (ignores other arguments)
		return NewLocalClientCreator(local)
	} else {
		switch proxy {
		// default mock applications
		case "mock://counter":
			return NewLocalClientCreator(counter.NewCounterApplication(false))
		case "mock://counter_serial":
			return NewLocalClientCreator(counter.NewCounterApplication(true))
		case "mock://kvstore":
			return NewLocalClientCreator(kvstore.NewKVStoreApplication())
		case "mock://persistent_kvstore":
			return NewLocalClientCreator(kvstore.NewPersistentKVStoreApplication(dbDir))
		case "mock://noop":
			return NewLocalClientCreator(abci.NewBaseApplication())
		default:
			// socket transport applications
			panic("proxy scheme not yet supported: " + proxy)
		}
	}
}
