package core

import (
	"fmt"
	"log/slog"

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
	NodeInfo() p2p.NodeInfo
}

type peers interface {
	AddPersistentPeers([]string) error
	DialPeersAsync([]string) error
	NumPeers() (outbound, inbound, dialig int)
	Peers() p2p.IPeerSet
}

// ----------------------------------------------
// These package level globals come with setters
// that are expected to be called only once, on startup

var (
	// external, thread safe interfaces
	proxyAppQuery appconn.Query

	// interfaces defined in types and above
	stateDB        dbm.DB
	blockStore     sm.BlockStore
	consensusState Consensus
	p2pPeers       peers
	p2pTransport   transport

	// objects
	pubKey        crypto.PubKey
	genDoc        *types.GenesisDoc // cache the genesis structure
	evsw          events.EventSwitch
	gTxDispatcher *txDispatcher
	mempool       mempl.Mempool
	getFastSync   func() bool // avoids dependency on consensus pkg

	logger *slog.Logger

	config cfg.RPCConfig
)

func SetStateDB(db dbm.DB) {
	stateDB = db
}

func SetBlockStore(bs sm.BlockStore) {
	blockStore = bs
}

func SetMempool(mem mempl.Mempool) {
	mempool = mem
}

func SetConsensusState(cs Consensus) {
	consensusState = cs
}

func SetP2PPeers(p peers) {
	p2pPeers = p
}

func SetP2PTransport(t transport) {
	p2pTransport = t
}

func SetPubKey(pk crypto.PubKey) {
	pubKey = pk
}

func SetGenesisDoc(doc *types.GenesisDoc) {
	genDoc = doc
}

func SetProxyAppQuery(appConn appconn.Query) {
	proxyAppQuery = appConn
}

func SetGetFastSync(v func() bool) {
	getFastSync = v
}

func SetLogger(l *slog.Logger) {
	logger = l
}

func SetEventSwitch(sw events.EventSwitch) {
	evsw = sw
	gTxDispatcher = newTxDispatcher(evsw)
}

func Start() {
	gTxDispatcher.Start()
}

// SetConfig sets an RPCConfig.
func SetConfig(c cfg.RPCConfig) {
	config = c
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
