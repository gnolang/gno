package core

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/rpc/config"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

const (
	// see README
	defaultPerPage = 30
	maxPerPage     = 100
)

// ----------------------------------------------
// These interfaces are used by RPC and must be thread safe

type Consensus interface {
	GetConfigDeepCopy() *cnscfg.ConsensusConfig
	GetState() sm.State
	GetValidators() (int64, []*types.Validator)
	GetLastHeight() int64
	GetRoundStateDeepCopy() *cstypes.RoundState
	GetRoundStateSimple() cstypes.RoundStateSimple
}

type transport interface {
	Listeners() []string
	IsListening() bool
	NodeInfo() p2pTypes.NodeInfo
}

type peers interface {
	Peers() p2p.PeerSet
}

// ----------------------------------------------
// Environment holds all per-node state that RPC handlers operate on.
// One Environment is created per Node instance, replacing the package-level
// globals this package used previously. All fields are expected to be
// populated before Start is called; individual handlers may only need a
// subset (tests construct partial Environments).
type Environment struct {
	// external, thread-safe interfaces
	ProxyAppQuery appconn.Query

	// interfaces defined in types and above
	StateDB      dbm.DB
	BlockStore   sm.BlockStore
	Consensus    Consensus
	P2PPeers     peers
	P2PTransport transport

	// objects
	PubKey      crypto.PubKey
	GenDoc      *types.GenesisDoc // cache the genesis structure
	EventSwitch events.EventSwitch
	Mempool     mempl.Mempool
	GetFastSync func() bool // avoids dependency on consensus pkg

	Logger *slog.Logger
	Config cfg.RPCConfig // value, not pointer — TimeoutBroadcastTxCommit must be stable

	// Internal state. Populated by Start; nil before.
	mtx          sync.Mutex
	txDispatcher *txDispatcher
	started      bool
	stopped      bool
}

// Start initializes any per-Environment background services. Currently this
// only creates and starts the txDispatcher if EventSwitch is non-nil.
// Start is idempotent but panics if called after Stop.
func (env *Environment) Start() error {
	env.mtx.Lock()
	defer env.mtx.Unlock()
	if env.stopped {
		panic("cannot Start a stopped Environment")
	}
	if env.started {
		return nil
	}
	if env.EventSwitch != nil {
		env.txDispatcher = newTxDispatcher(env.EventSwitch)
	}
	env.started = true
	return nil
}

// Stop tears down the services started by Start. It should be called before
// the associated EventSwitch is stopped so the txDispatcher goroutine exits
// via its own Quit channel rather than racing evsw.Quit(). Stop is idempotent.
func (env *Environment) Stop() error {
	env.mtx.Lock()
	defer env.mtx.Unlock()
	if env.stopped {
		return nil
	}
	env.stopped = true
	if env.txDispatcher != nil && env.txDispatcher.IsRunning() {
		if err := env.txDispatcher.Stop(); err != nil {
			panic(fmt.Sprintf("txDispatcher.Stop: %v", err))
		}
	}
	env.started = false
	return nil
}

func validatePage(page, perPage, totalCount int) (int, error) {
	if perPage < 1 {
		panic(fmt.Sprintf("zero or negative perPage: %d", perPage))
	}

	if page == 0 {
		return 1, nil // default
	}

	pages := ((totalCount - 1) / perPage) + 1
	if pages == 0 {
		pages = 1 // one page (even if it's empty)
	}
	if page < 0 || page > pages {
		return 1, fmt.Errorf("page should be within [0, %d] range, given %d", pages, page)
	}

	return page, nil
}

func validatePerPage(perPage int) int {
	if perPage < 1 {
		return defaultPerPage
	} else if perPage > maxPerPage {
		return maxPerPage
	}
	return perPage
}
