package proxy

import (
	"runtime"
	"sync"
	"time"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/counter"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/meminfo"
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

// LocalApp returns the in-process application the clients call.
func (l *localClientCreator) LocalApp() abci.Application { return l.app }

// NewReadOnlyABCIClient returns a client whose calls do not contend with
// consensus or mempool, and do not serialise against each other: as many of
// them as newMemoryQueryLimiter admits may be inside the application at once.
//
// This rests on an invariant: everything reachable from Application.Query and
// Application.Info is goroutine-safe (immutable DB snapshots, SyncGoMap
// cacheNodes sealed on publication, atomic last-block header, per-tx forked
// allocator). Breaking it reintroduces a data race.
//
// The returned client satisfies the full abcicli.Client interface, but only the
// three methods appconn.Query exposes — EchoSync, InfoSync, QuerySync — are safe
// to use on it. The one that cannot merely be documented is SetResponseCallback,
// which readOnlyClient rejects.
func (l *localClientCreator) NewReadOnlyABCIClient() (abcicli.Client, error) {
	cli := abcicli.NewLocalClient(newMemoryQueryLimiter(meminfo.Read), l.app)
	return readOnlyClient{Client: cli}, nil
}

// readOnlyClient forwards everything to the bounded localClient underneath and
// rejects the one method the bound makes unsafe.
//
// localClient guards SetResponseCallback with the same Locker it holds for every
// call into Application. On the consensus and mempool connections that Locker is
// a mutex, so the write to Callback is genuinely excluded from completeRequest's
// concurrent reads; under queryLimiter the Lock takes one slot of N and excludes
// nothing. Documenting that leaves a Lock/Unlock pair that reads as
// synchronisation and provides none, so the method is refused here instead.
//
// The remaining mutating methods stay callable and documented rather than
// guarded, as the ADR records: nothing reaches them through appconn.Query, and
// each needs its own override. This is the place to add them if a caller ever
// appears.
type readOnlyClient struct {
	abcicli.Client
}

// SetResponseCallback panics: see readOnlyClient.
func (readOnlyClient) SetResponseCallback(abcicli.Callback) {
	panic("proxy: SetResponseCallback on the read-only ABCI connection: the " +
		"connection admits concurrent callers, so the write would race with " +
		"completeRequest's reads; use a client from NewABCIClient")
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
// The bound moves between min and max with the machine's free memory, the way
// the integration suite's nodeBudget does: a quarter of the memory is left to
// everything else, and the limit goes up one slot while callers fill every slot
// and the free memory covers that reserve plus queryMemMargin, down one while it
// falls short of the reserve. The margin between the two thresholds is what
// keeps the limit from oscillating.
//
// A limit of 1 makes this exactly the mutex it replaced.
type queryLimiter struct {
	mu      sync.Mutex
	freed   *sync.Cond // signalled on Unlock
	running int
	waiting int // callers blocked in Lock

	// limit ratchets between min and max; see resize.
	limit, min, max int
	lastCheck       time.Time

	readMem func() (meminfo.Info, bool)
}

// queryMemReadInterval paces memory readings, so that a burst of callers
// deciding at once see one figure between them rather than each growing the
// limit against it. It also bounds how fast the limit moves: a slot a second.
const queryMemReadInterval = time.Second

// queryMemReserveDiv leaves a quarter of the memory this process may use to
// everything else, as nodeBudget's nodeMemReserveDiv does. It absorbs the
// memory running queries have yet to take, which resize does not count.
const queryMemReserveDiv = 4

// queryMemMargin is the room above the reserve that a further slot needs: one
// more query to grow into. A query's real cost is the application's business,
// so this is a margin rather than a measurement, set well above what a typical
// query holds.
const queryMemMargin = 1 << 30

// newQueryLimiter returns a limiter fixed at n slots, at least one.
func newQueryLimiter(n int) *queryLimiter {
	n = max(n, 1)
	l := &queryLimiter{limit: n, min: n, max: n}
	l.freed = sync.NewCond(&l.mu)
	return l
}

// queryLimiterFloor is where the limit starts: the count CI validates, as
// testutils' defaultMaxParallel is.
const queryLimiterFloor = 4

// newMemoryQueryLimiter returns a limiter that ramps from
// min(GOMAXPROCS, queryLimiterFloor) towards GOMAXPROCS as readMem allows. Query
// work is CPU-bound in the VM, so there is nothing to gain past GOMAXPROCS,
// which is read here rather than at package init: since Go 1.25 its default
// tracks the cgroup CPU limit and moves when that limit does.
//
// Where memory cannot be read the limiter is fixed at GOMAXPROCS, the bound
// before the ramp existed: with no reading to go on, the machine's cores are
// the only measure of what it can handle.
func newMemoryQueryLimiter(readMem func() (meminfo.Info, bool)) *queryLimiter {
	most := runtime.GOMAXPROCS(0)
	floor := min(most, queryLimiterFloor)
	if _, ok := readMem(); !ok || floor == most {
		return newQueryLimiter(most)
	}
	l := newQueryLimiter(floor)
	l.max = most
	l.readMem = readMem
	return l
}

// Lock blocks until a slot is free. localClient defers every Unlock, so a slot
// is returned even when the application panics.
func (q *queryLimiter) Lock() {
	q.mu.Lock()
	defer q.mu.Unlock()
	for {
		q.resize()
		if q.running < q.limit {
			q.running++
			return
		}
		q.waiting++
		q.freed.Wait()
		q.waiting--
	}
}

func (q *queryLimiter) Unlock() {
	q.mu.Lock()
	q.running--
	q.mu.Unlock()
	q.freed.Signal()
}

// resize moves the limit one step towards what the machine can take. Called
// with q.mu held, from Lock only: a slot that a grown limit adds goes to the
// caller that grew it or to a waiter it wakes, and a limit that shrinks takes
// effect as running queries return.
//
// Unlike nodeBudget it does not count memory a running query has yet to take;
// the reserve absorbs that overshoot.
func (q *queryLimiter) resize() {
	now := time.Now()
	if q.readMem == nil || now.Sub(q.lastCheck) < queryMemReadInterval {
		return
	}
	// Paced whether or not the reading succeeds, so a failing one never puts
	// file reads on every query.
	q.lastCheck = now
	mi, ok := q.readMem()
	if !ok {
		return
	}
	reserve := mi.Total / queryMemReserveDiv
	switch {
	case mi.Available < reserve && q.limit > q.min:
		q.limit--
	case mi.Available >= reserve+queryMemMargin && q.limit < q.max && q.running+q.waiting >= q.limit:
		// Only when the current limit is saturated: raising it while slots sit
		// free buys nothing. Callers waiting for a slot count, since the reading
		// after an Unlock lands in the waiter it woke, which always finds a slot
		// free.
		q.limit++
		// The caller may be taking the slot Unlock freed, which leaves the new
		// one to a waiter.
		q.freed.Signal()
	}
}

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
