package proxy

import (
	"sync"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/counter"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

// ClientCreator creates ABCI clients for the three Tendermint connections.
type ClientCreator interface {
	// NewABCIClient returns a client for mutating connections (consensus, mempool).
	NewABCIClient() (abcicli.Client, error)
	// NewReadOnlyABCIClient returns a client for the query connection.
	// It uses an independent mutex so query calls never block consensus.
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

// NewReadOnlyABCIClient returns a client that takes no lock at all, so query
// calls neither contend with consensus and mempool nor with each other.
//
// This rests on an invariant: everything reachable from Application.Query and
// Application.Info is goroutine-safe (immutable DB snapshots, SyncGoMap
// cacheNodes, atomic last-block header, per-tx forked allocator). Breaking it
// reintroduces a data race.
//
// The returned client satisfies the full abcicli.Client interface, but only the
// read-only subset of appconn.Query (Echo, Info, Query) is safe to use on it.
// In particular SetResponseCallback is unsynchronised here: its write to the
// client's Callback field would race with the concurrent reads in
// completeRequest. Nothing calls it on this connection today — the appconn
// Query wrapper does not expose it — and nothing should.
func (l *localClientCreator) NewReadOnlyABCIClient() (abcicli.Client, error) {
	return abcicli.NewLocalClient(noLock, l.app), nil
}

// noLock is the Locker handed to read-only connections. It is stateless, so one
// shared value serves every client.
var noLock sync.Locker = noopMutex{}

type noopMutex struct{}

func (noopMutex) Lock()   {}
func (noopMutex) Unlock() {}

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
