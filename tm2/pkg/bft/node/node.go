package node

// Ignore pprof import for gosec, as profiling
// is enabled by the user by setting a profiling address

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/file"
	"github.com/rs/cors"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bc "github.com/gnolang/gno/tm2/pkg/bft/blockchain"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	cs "github.com/gnolang/gno/tm2/pkg/bft/consensus"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	rpccore "github.com/gnolang/gno/tm2/pkg/bft/rpc/core"
	_ "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpcserver "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/null"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/bft/version"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/service"
	verset "github.com/gnolang/gno/tm2/pkg/versionset"
)

// ------------------------------------------------------------------------------

// DBContext specifies config information for loading a new DB.
type DBContext struct {
	ID     string
	Config *cfg.Config
}

// DBProvider takes a DBContext and returns an instantiated DB.
type DBProvider func(*DBContext) (dbm.DB, error)

// DefaultDBProvider returns a database using the db.Backend and DBDir
// specified in the ctx.Config.
func DefaultDBProvider(ctx *DBContext) (dbm.DB, error) {
	dbType := dbm.BackendType(ctx.Config.DBBackend)
	return dbm.NewDB(ctx.ID, dbType, ctx.Config.DBDir())
}

// GenesisDocProvider returns a GenesisDoc.
// It allows the GenesisDoc to be pulled from sources other than the
// filesystem, for instance from a distributed key-value store cluster.
type GenesisDocProvider func() (*types.GenesisDoc, error)

// DefaultGenesisDocProviderFunc returns a GenesisDocProvider that loads
// the GenesisDoc from the genesis path on the filesystem.
func DefaultGenesisDocProviderFunc(genesisFile string) GenesisDocProvider {
	return func() (*types.GenesisDoc, error) {
		return types.GenesisDocFromFile(genesisFile)
	}
}

// NodeProvider takes a config and a logger and returns a ready to go Node.
type NodeProvider func(*cfg.Config, *slog.Logger) (*Node, error)

// DefaultNewNode returns a Tendermint node with default settings for the
// PrivValidator, ClientCreator, GenesisDoc, and DBProvider.
// It implements NodeProvider.
func DefaultNewNode(config *cfg.Config, genesisFile string, logger *slog.Logger) (*Node, error) {
	// Generate node PrivKey
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	// Get privValKeys.
	newPrivValKey := config.PrivValidatorKeyFile()
	newPrivValState := config.PrivValidatorStateFile()

	// Get app client creator.
	appClientCreator := proxy.DefaultClientCreator(
		config.LocalApp,
		config.ProxyApp,
		config.ABCI,
		config.DBDir(),
	)

	return NewNode(config,
		privval.LoadOrGenFilePV(newPrivValKey, newPrivValState),
		nodeKey,
		appClientCreator,
		DefaultGenesisDocProviderFunc(genesisFile),
		DefaultDBProvider,
		logger,
	)
}

// Option sets a parameter for the node.
type Option func(*Node)

// CustomReactors allows you to add custom reactors (name -> p2p.Reactor) to
// the node's Switch.
//
// WARNING: using any name from the below list of the existing reactors will
// result in replacing it with the custom one.
//
//   - MEMPOOL
//   - BLOCKCHAIN
//   - CONSENSUS
//   - EVIDENCE
//   - PEX
func CustomReactors(reactors map[string]p2p.Reactor) Option {
	return func(n *Node) {
		for name, reactor := range reactors {
			if existingReactor := n.sw.Reactor(name); existingReactor != nil {
				n.sw.Logger.Info("Replacing existing reactor with a custom one",
					"name", name, "existing", existingReactor, "custom", reactor)
				n.sw.RemoveReactor(name, existingReactor)
			}
			n.sw.AddReactor(name, reactor)
		}
	}
}

// ------------------------------------------------------------------------------

// Node is the highest level interface to a full Tendermint node.
// It includes all configuration information and running services.
type Node struct {
	service.BaseService

	// config
	config        *cfg.Config
	genesisDoc    *types.GenesisDoc   // initial validator set
	privValidator types.PrivValidator // local node's validator key

	// network
	transport   *p2p.MultiplexTransport
	sw          *p2p.Switch // p2p connections
	nodeInfo    p2p.NodeInfo
	nodeKey     *p2p.NodeKey // our node privkey
	isListening bool

	// services
	evsw              events.EventSwitch
	stateDB           dbm.DB
	blockStore        *store.BlockStore // store the blockchain to disk
	bcReactor         p2p.Reactor       // for fast-syncing
	mempoolReactor    *mempl.Reactor    // for gossipping transactions
	mempool           mempl.Mempool
	consensusState    *cs.ConsensusState   // latest consensus state
	consensusReactor  *cs.ConsensusReactor // for participating in the consensus
	proxyApp          appconn.AppConns     // connection to the application
	rpcListeners      []net.Listener       // rpc servers
	txEventStore      eventstore.TxEventStore
	eventStoreService *eventstore.Service
	firstBlockSignal  <-chan struct{}
}

func initDBs(config *cfg.Config, dbProvider DBProvider) (blockStore *store.BlockStore, stateDB dbm.DB, err error) {
	var blockStoreDB dbm.DB
	blockStoreDB, err = dbProvider(&DBContext{"blockstore", config})
	if err != nil {
		return
	}
	blockStore = store.NewBlockStore(blockStoreDB)

	stateDB, err = dbProvider(&DBContext{"state", config})
	if err != nil {
		return
	}

	return
}

func createAndStartProxyAppConns(clientCreator proxy.ClientCreator, logger *slog.Logger) (appconn.AppConns, error) {
	proxyApp := appconn.NewAppConns(clientCreator)
	proxyApp.SetLogger(logger.With("module", "proxy"))
	if err := proxyApp.Start(); err != nil {
		return nil, fmt.Errorf("error starting proxy app connections: %w", err)
	}
	return proxyApp, nil
}

func createAndStartEventStoreService(
	cfg *cfg.Config,
	evsw events.EventSwitch,
	logger *slog.Logger,
) (*eventstore.Service, eventstore.TxEventStore, error) {
	var (
		err          error
		txEventStore eventstore.TxEventStore
	)

	// Instantiate the event store based on the configuration
	switch cfg.TxEventStore.EventStoreType {
	case file.EventStoreType:
		// Transaction events should be logged to files
		txEventStore, err = file.NewTxEventStore(cfg.TxEventStore)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to create file tx event store, %w", err)
		}
	default:
		// Transaction event storing should be omitted
		txEventStore = null.NewNullEventStore()
	}

	indexerService := eventstore.NewEventStoreService(txEventStore, evsw)
	indexerService.SetLogger(logger.With("module", "eventstore"))
	if err := indexerService.Start(); err != nil {
		return nil, nil, err
	}

	return indexerService, txEventStore, nil
}

func doHandshake(stateDB dbm.DB, state sm.State, blockStore sm.BlockStore,
	genDoc *types.GenesisDoc, evsw events.EventSwitch, proxyApp appconn.AppConns, consensusLogger *slog.Logger,
) error {
	handshaker := cs.NewHandshaker(stateDB, state, blockStore, genDoc)
	handshaker.SetLogger(consensusLogger)
	handshaker.SetEventSwitch(evsw)
	if err := handshaker.Handshake(proxyApp); err != nil {
		return fmt.Errorf("error during handshake: %w", err)
	}
	return nil
}

func logNodeStartupInfo(state sm.State, pubKey crypto.PubKey, logger, consensusLogger *slog.Logger) {
	// Log the version info.
	logger.Info("Version info",
		"version", version.Version,
	)

	addr := pubKey.Address()
	// Log whether this node is a validator or an observer
	if state.Validators.HasAddress(addr) {
		consensusLogger.Info("This node is a validator", "addr", addr, "pubKey", pubKey)
	} else {
		consensusLogger.Info("This node is not a validator", "addr", addr, "pubKey", pubKey)
	}
}

func onlyValidatorIsUs(state sm.State, privVal types.PrivValidator) bool {
	if state.Validators.Size() > 1 {
		return false
	}
	addr, _ := state.Validators.GetByIndex(0)
	return privVal.GetPubKey().Address() == addr
}

func createMempoolAndMempoolReactor(config *cfg.Config, proxyApp appconn.AppConns,
	state sm.State, logger *slog.Logger,
) (*mempl.Reactor, *mempl.CListMempool) {
	mempool := mempl.NewCListMempool(
		config.Mempool,
		proxyApp.Mempool(),
		state.LastBlockHeight,
		state.ConsensusParams.Block.MaxTxBytes,
		mempl.WithPreCheck(sm.TxPreCheck(state)),
	)
	mempoolLogger := logger.With("module", "mempool")
	mempoolReactor := mempl.NewReactor(config.Mempool, mempool)
	mempoolReactor.SetLogger(mempoolLogger)

	if config.Consensus.WaitForTxs() {
		mempool.EnableTxsAvailable()
	}
	return mempoolReactor, mempool
}

func createBlockchainReactor(config *cfg.Config,
	state sm.State,
	blockExec *sm.BlockExecutor,
	blockStore *store.BlockStore,
	fastSync bool,
	logger *slog.Logger,
) (bcReactor p2p.Reactor, err error) {
	bcReactor = bc.NewBlockchainReactor(state.Copy(), blockExec, blockStore, fastSync)

	bcReactor.SetLogger(logger.With("module", "blockchain"))
	return bcReactor, nil
}

func createConsensusReactor(config *cfg.Config,
	state sm.State,
	blockExec *sm.BlockExecutor,
	blockStore sm.BlockStore,
	mempool *mempl.CListMempool,
	privValidator types.PrivValidator,
	fastSync bool,
	evsw events.EventSwitch,
	consensusLogger *slog.Logger,
) (*cs.ConsensusReactor, *cs.ConsensusState) {
	consensusState := cs.NewConsensusState(
		config.Consensus,
		state.Copy(),
		blockExec,
		blockStore,
		mempool,
	)
	consensusState.SetLogger(consensusLogger)
	if privValidator != nil {
		consensusState.SetPrivValidator(privValidator)
	}
	consensusReactor := cs.NewConsensusReactor(consensusState, fastSync)
	consensusReactor.SetLogger(consensusLogger)
	consensusReactor.SetEventSwitch(evsw)
	// services which will be publishing and/or subscribing for messages (events)
	// consensusReactor will set it on consensusState and blockExecutor
	return consensusReactor, consensusState
}

func createTransport(config *cfg.Config, nodeInfo p2p.NodeInfo, nodeKey *p2p.NodeKey, proxyApp appconn.AppConns) (*p2p.MultiplexTransport, []p2p.PeerFilterFunc) {
	var (
		mConnConfig = p2p.MConnConfig(config.P2P)
		transport   = p2p.NewMultiplexTransport(nodeInfo, *nodeKey, mConnConfig)
		connFilters = []p2p.ConnFilterFunc{}
		peerFilters = []p2p.PeerFilterFunc{}
	)

	if !config.P2P.AllowDuplicateIP {
		connFilters = append(connFilters, p2p.ConnDuplicateIPFilter())
	}

	// Filter peers by addr or pubkey with an ABCI query.
	// If the query return code is OK, add peer.
	if config.FilterPeers {
		connFilters = append(
			connFilters,
			// ABCI query for address filtering.
			func(_ p2p.ConnSet, c net.Conn, _ []net.IP) error {
				res, err := proxyApp.Query().QuerySync(abci.RequestQuery{
					Path: fmt.Sprintf("/p2p/filter/addr/%s", c.RemoteAddr().String()),
				})
				if err != nil {
					return err
				}
				if res.IsErr() {
					return fmt.Errorf("error querying abci app: %v", res)
				}

				return nil
			},
		)

		peerFilters = append(
			peerFilters,
			// ABCI query for ID filtering.
			func(_ p2p.IPeerSet, p p2p.Peer) error {
				res, err := proxyApp.Query().QuerySync(abci.RequestQuery{
					Path: fmt.Sprintf("/p2p/filter/id/%s", p.ID()),
				})
				if err != nil {
					return err
				}
				if res.IsErr() {
					return fmt.Errorf("error querying abci app: %v", res)
				}

				return nil
			},
		)
	}

	p2p.MultiplexTransportConnFilters(connFilters...)(transport)
	return transport, peerFilters
}

func createSwitch(config *cfg.Config,
	transport *p2p.MultiplexTransport,
	peerFilters []p2p.PeerFilterFunc,
	mempoolReactor *mempl.Reactor,
	bcReactor p2p.Reactor,
	consensusReactor *cs.ConsensusReactor,
	nodeInfo p2p.NodeInfo,
	nodeKey *p2p.NodeKey,
	p2pLogger *slog.Logger,
) *p2p.Switch {
	sw := p2p.NewSwitch(
		config.P2P,
		transport,
		p2p.SwitchPeerFilters(peerFilters...),
	)
	sw.SetLogger(p2pLogger)
	sw.AddReactor("MEMPOOL", mempoolReactor)
	sw.AddReactor("BLOCKCHAIN", bcReactor)
	sw.AddReactor("CONSENSUS", consensusReactor)

	sw.SetNodeInfo(nodeInfo)
	sw.SetNodeKey(nodeKey)

	p2pLogger.Info("P2P Node ID", "ID", nodeKey.ID(), "file", config.NodeKeyFile())
	return sw
}

// NewNode returns a new, ready to go, Tendermint Node.
func NewNode(config *cfg.Config,
	privValidator types.PrivValidator,
	nodeKey *p2p.NodeKey,
	clientCreator appconn.ClientCreator,
	genesisDocProvider GenesisDocProvider,
	dbProvider DBProvider,
	logger *slog.Logger,
	options ...Option,
) (*Node, error) {
	blockStore, stateDB, err := initDBs(config, dbProvider)
	if err != nil {
		return nil, err
	}

	state, genDoc, err := LoadStateFromDBOrGenesisDocProvider(stateDB, genesisDocProvider)
	if err != nil {
		return nil, err
	}

	// Create the proxyApp and establish connections to the ABCI app (consensus, mempool, query).
	proxyApp, err := createAndStartProxyAppConns(clientCreator, logger)
	if err != nil {
		return nil, err
	}

	// EventSwitch and EventStoreService must be started before the handshake because
	// we might need to store the txs of the replayed block as this might not have happened
	// when the node stopped last time (i.e. the node stopped after it saved the block
	// but before it indexed the txs, or, endblocker panicked)
	evsw := events.NewEventSwitch()

	// Signal readiness when receiving the first block.
	const readinessListenerID = "first_block_listener"

	cFirstBlock := make(chan struct{})
	var once sync.Once
	evsw.AddListener(readinessListenerID, func(ev events.Event) {
		if _, ok := ev.(types.EventNewBlock); ok {
			once.Do(func() {
				close(cFirstBlock)
				evsw.RemoveListener(readinessListenerID)
			})
		}
	})

	// Transaction event storing
	eventStoreService, txEventStore, err := createAndStartEventStoreService(config, evsw, logger)
	if err != nil {
		return nil, err
	}

	// Create the handshaker, which calls RequestInfo, sets the AppVersion on the state,
	// and replays any blocks as necessary to sync tendermint with the app.
	consensusLogger := logger.With("module", "consensus")
	if err := doHandshake(stateDB, state, blockStore, genDoc, evsw, proxyApp, consensusLogger); err != nil {
		return nil, err
	}

	// Reload the state. It will have the Version.Consensus.App set by the
	// Handshake, and may have other modifications as well (ie. depending on
	// what happened during block replay).
	state = sm.LoadState(stateDB)

	// If an address is provided, listen on the socket for a connection from an
	// external signing process.
	if config.PrivValidatorListenAddr != "" {
		// FIXME: we should start services inside OnStart
		privValidator, err = createAndStartPrivValidatorSocketClient(config.PrivValidatorListenAddr, logger)
		if err != nil {
			return nil, errors.Wrap(err, "error with private validator socket client")
		}
	}

	pubKey := privValidator.GetPubKey()
	if pubKey == nil {
		// TODO: GetPubKey should return errors - https://github.com/gnolang/gno/tm2/pkg/bft/issues/3602
		return nil, errors.New("could not retrieve public key from private validator")
	}

	logNodeStartupInfo(state, pubKey, logger, consensusLogger)

	// Decide whether to fast-sync or not
	// We don't fast-sync when the only validator is us.
	fastSync := config.FastSyncMode && !onlyValidatorIsUs(state, privValidator)

	// Make MempoolReactor
	mempoolReactor, mempool := createMempoolAndMempoolReactor(config, proxyApp, state, logger)

	// make block executor for consensus and blockchain reactors to execute blocks
	blockExec := sm.NewBlockExecutor(
		stateDB,
		logger.With("module", "state"),
		proxyApp.Consensus(),
		mempool,
	)

	// Make BlockchainReactor
	bcReactor, err := createBlockchainReactor(config, state, blockExec, blockStore, fastSync, logger)
	if err != nil {
		return nil, errors.Wrap(err, "could not create blockchain reactor")
	}

	// Make ConsensusReactor
	consensusReactor, consensusState := createConsensusReactor(
		config, state, blockExec, blockStore, mempool,
		privValidator, fastSync, evsw, consensusLogger,
	)

	nodeInfo, err := makeNodeInfo(config, nodeKey, txEventStore, genDoc, state)
	if err != nil {
		return nil, errors.Wrap(err, "error making NodeInfo")
	}

	// Setup Transport.
	transport, peerFilters := createTransport(config, nodeInfo, nodeKey, proxyApp)

	// Setup Switch.
	p2pLogger := logger.With("module", "p2p")
	sw := createSwitch(
		config, transport, peerFilters, mempoolReactor, bcReactor,
		consensusReactor, nodeInfo, nodeKey, p2pLogger,
	)

	err = sw.AddPersistentPeers(splitAndTrimEmpty(config.P2P.PersistentPeers, ",", " "))
	if err != nil {
		return nil, errors.Wrap(err, "could not add peers from persistent_peers field")
	}

	if config.ProfListenAddress != "" {
		server := &http.Server{
			Addr:              config.ProfListenAddress,
			ReadHeaderTimeout: 60 * time.Second,
		}

		go func() {
			logger.Error("Profile server", "err", server.ListenAndServe())
		}()
	}

	node := &Node{
		config:        config,
		genesisDoc:    genDoc,
		privValidator: privValidator,

		transport: transport,
		sw:        sw,
		nodeInfo:  nodeInfo,
		nodeKey:   nodeKey,

		evsw:              evsw,
		stateDB:           stateDB,
		blockStore:        blockStore,
		bcReactor:         bcReactor,
		mempoolReactor:    mempoolReactor,
		mempool:           mempool,
		consensusState:    consensusState,
		consensusReactor:  consensusReactor,
		proxyApp:          proxyApp,
		txEventStore:      txEventStore,
		eventStoreService: eventStoreService,
		firstBlockSignal:  cFirstBlock,
	}
	node.BaseService = *service.NewBaseService(logger, "Node", node)

	for _, option := range options {
		option(node)
	}

	return node, nil
}

// OnStart starts the Node. It implements service.Service.
func (n *Node) OnStart() error {
	now := tmtime.Now()
	genTime := n.genesisDoc.GenesisTime
	if genTime.After(now) {
		n.Logger.Info("Genesis time is in the future. Sleeping until then...", "genTime", genTime)
		time.Sleep(genTime.Sub(now))
	}

	// Set up the GLOBAL variables in rpc/core which refer to this node.
	// This is done separately from startRPC(), as the values in rpc/core are used,
	// for instance, to set up Local clients (rpc/client) which work without
	// a network connection.
	n.configureRPC()
	if n.config.RPC.Unsafe {
		rpccore.AddUnsafeRoutes()
	}
	rpccore.Start()

	// Start the RPC server before the P2P server
	// so we can eg. receive txs for the first block
	if n.config.RPC.ListenAddress != "" {
		listeners, err := n.startRPC()
		if err != nil {
			return err
		}
		n.rpcListeners = listeners
	}

	// Start the transport.
	addr, err := p2p.NewNetAddressFromString(p2p.NetAddressString(n.nodeKey.ID(), n.config.P2P.ListenAddress))
	if err != nil {
		return err
	}
	if err := n.transport.Listen(*addr); err != nil {
		return err
	}
	if addr.Port == 0 {
		// if the port we have from config.P2p.ListenAdress is 0,
		// it means the port was selected when doing net.Listen (using autoselect on the kernel).
		// fix the config variable using the correct address
		na := n.transport.NetAddress()
		n.config.P2P.ListenAddress = na.DialString()
	}

	n.isListening = true

	if n.config.Mempool.WalEnabled() {
		n.mempool.InitWAL() // no need to have the mempool wal during tests
	}

	// Start the switch (the P2P server).
	err = n.sw.Start()
	if err != nil {
		return err
	}

	// Always connect to persistent peers
	err = n.sw.DialPeersAsync(splitAndTrimEmpty(n.config.P2P.PersistentPeers, ",", " "))
	if err != nil {
		return errors.Wrap(err, "could not dial peers from persistent_peers field")
	}

	return nil
}

// OnStop stops the Node. It implements service.Service.
func (n *Node) OnStop() {
	n.BaseService.OnStop()

	n.Logger.Info("Stopping Node")

	// first stop the non-reactor services
	n.evsw.Stop()
	n.eventStoreService.Stop()

	// now stop the reactors
	n.sw.Stop()

	// stop mempool WAL
	if n.config.Mempool.WalEnabled() {
		n.mempool.CloseWAL()
	}

	n.isListening = false

	// finally stop the listeners / external services
	for _, l := range n.rpcListeners {
		n.Logger.Info("Closing rpc listener", "listener", l)
		if err := l.Close(); err != nil {
			n.Logger.Error("Error closing listener", "listener", l, "err", err)
		}
	}

	if pvsc, ok := n.privValidator.(service.Service); ok {
		pvsc.Stop()
	}
}

// Ready signals that the node is ready by returning a blocking channel. This channel is closed when the node receives its first block.
func (n *Node) Ready() <-chan struct{} {
	return n.firstBlockSignal
}

// configureRPC sets all variables in rpccore so they will serve
// rpc calls from this node
func (n *Node) configureRPC() {
	rpccore.SetStateDB(n.stateDB)
	rpccore.SetBlockStore(n.blockStore)
	rpccore.SetConsensusState(n.consensusState)
	rpccore.SetMempool(n.mempool)
	rpccore.SetP2PPeers(n.sw)
	rpccore.SetP2PTransport(n)
	pubKey := n.privValidator.GetPubKey()
	rpccore.SetPubKey(pubKey)
	rpccore.SetGenesisDoc(n.genesisDoc)
	rpccore.SetProxyAppQuery(n.proxyApp.Query())
	rpccore.SetGetFastSync(n.consensusReactor.FastSync)
	rpccore.SetLogger(n.Logger.With("module", "rpc"))
	rpccore.SetEventSwitch(n.evsw)
	rpccore.SetConfig(*n.config.RPC)
}

func (n *Node) startRPC() ([]net.Listener, error) {
	listenAddrs := splitAndTrimEmpty(n.config.RPC.ListenAddress, ",", " ")

	config := rpcserver.DefaultConfig()
	config.MaxBodyBytes = n.config.RPC.MaxBodyBytes
	config.MaxHeaderBytes = n.config.RPC.MaxHeaderBytes
	config.MaxOpenConnections = n.config.RPC.MaxOpenConnections
	// If necessary adjust global WriteTimeout to ensure it's greater than
	// TimeoutBroadcastTxCommit.
	// See https://github.com/gnolang/gno/tm2/pkg/bft/issues/3435
	if config.WriteTimeout <= n.config.RPC.TimeoutBroadcastTxCommit {
		config.WriteTimeout = n.config.RPC.TimeoutBroadcastTxCommit + 1*time.Second
	}

	// we may expose the rpc over both a unix and tcp socket
	var rebuildAddresses bool
	listeners := make([]net.Listener, len(listenAddrs))
	for i, listenAddr := range listenAddrs {
		mux := http.NewServeMux()
		rpcLogger := n.Logger.With("module", "rpc-server")
		wmLogger := rpcLogger.With("protocol", "websocket")
		wm := rpcserver.NewWebsocketManager(rpccore.Routes,
			rpcserver.OnDisconnect(func(remoteAddr string) {
				// any cleanup...
				// (we used to unsubscribe from all event subscriptions)
			}),
			rpcserver.ReadLimit(config.MaxBodyBytes),
		)
		wm.SetLogger(wmLogger)
		mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, rpccore.Routes, rpcLogger)
		if strings.HasPrefix(listenAddr, "tcp://") && strings.HasSuffix(listenAddr, ":0") {
			rebuildAddresses = true
		}
		listener, err := rpcserver.Listen(
			listenAddr,
			config,
		)
		if err != nil {
			return nil, err
		}

		var rootHandler http.Handler = mux
		if n.config.RPC.IsCorsEnabled() {
			corsMiddleware := cors.New(cors.Options{
				AllowedOrigins: n.config.RPC.CORSAllowedOrigins,
				AllowedMethods: n.config.RPC.CORSAllowedMethods,
				AllowedHeaders: n.config.RPC.CORSAllowedHeaders,
			})
			rootHandler = corsMiddleware.Handler(mux)
		}
		if n.config.RPC.IsTLSEnabled() {
			go rpcserver.StartHTTPAndTLSServer(
				listener,
				rootHandler,
				n.config.RPC.CertFile(),
				n.config.RPC.KeyFile(),
				rpcLogger,
				config,
			)
		} else {
			go rpcserver.StartHTTPServer(
				listener,
				rootHandler,
				rpcLogger,
				config,
			)
		}

		listeners[i] = listener
	}
	if rebuildAddresses {
		n.config.RPC.ListenAddress = joinListenerAddresses(listeners)
	}

	return listeners, nil
}

func joinListenerAddresses(ll []net.Listener) string {
	sl := make([]string, len(ll))
	for i, l := range ll {
		sl[i] = l.Addr().Network() + "://" + l.Addr().String()
	}
	return strings.Join(sl, ",")
}

// Switch returns the Node's Switch.
func (n *Node) Switch() *p2p.Switch {
	return n.sw
}

// EventSwitch returns the node's EventSwitch.
func (n *Node) EventSwitch() events.EventSwitch {
	return n.evsw
}

// BlockStore returns the Node's BlockStore.
func (n *Node) BlockStore() *store.BlockStore {
	return n.blockStore
}

// ConsensusState returns the Node's ConsensusState.
func (n *Node) ConsensusState() *cs.ConsensusState {
	return n.consensusState
}

// ConsensusReactor returns the Node's ConsensusReactor.
func (n *Node) ConsensusReactor() *cs.ConsensusReactor {
	return n.consensusReactor
}

// MempoolReactor returns the Node's mempool reactor.
func (n *Node) MempoolReactor() *mempl.Reactor {
	return n.mempoolReactor
}

// Mempool returns the Node's mempool.
func (n *Node) Mempool() mempl.Mempool {
	return n.mempool
}

// PrivValidator returns the Node's PrivValidator.
// XXX: for convenience only!
func (n *Node) PrivValidator() types.PrivValidator {
	return n.privValidator
}

// GenesisDoc returns the Node's GenesisDoc.
func (n *Node) GenesisDoc() *types.GenesisDoc {
	return n.genesisDoc
}

// ProxyApp returns the Node's AppConns, representing its connections to the ABCI application.
func (n *Node) ProxyApp() appconn.AppConns {
	return n.proxyApp
}

// Config returns the Node's config.
func (n *Node) Config() *cfg.Config {
	return n.config
}

// ------------------------------------------------------------------------------

func (n *Node) Listeners() []string {
	return []string{
		fmt.Sprintf("Listener(@%v)", n.config.P2P.ExternalAddress),
	}
}

func (n *Node) IsListening() bool {
	return n.isListening
}

// NodeInfo returns the Node's Info from the Switch.
func (n *Node) NodeInfo() p2p.NodeInfo {
	return n.nodeInfo
}

func makeNodeInfo(
	config *cfg.Config,
	nodeKey *p2p.NodeKey,
	txEventStore eventstore.TxEventStore,
	genDoc *types.GenesisDoc,
	state sm.State,
) (p2p.NodeInfo, error) {
	txIndexerStatus := eventstore.StatusOff
	if txEventStore.GetType() != null.EventStoreType {
		txIndexerStatus = eventstore.StatusOn
	}

	bcChannel := bc.BlockchainChannel
	vset := version.VersionSet
	vset.Set(verset.VersionInfo{
		Name:    "app",
		Version: state.AppVersion,
	})

	nodeInfo := p2p.NodeInfo{
		VersionSet: vset,
		Network:    genDoc.ChainID,
		Version:    version.Version,
		Channels: []byte{
			bcChannel,
			cs.StateChannel, cs.DataChannel, cs.VoteChannel, cs.VoteSetBitsChannel,
			mempl.MempoolChannel,
		},
		Moniker: config.Moniker,
		Other: p2p.NodeInfoOther{
			TxIndex:    txIndexerStatus,
			RPCAddress: config.RPC.ListenAddress,
		},
	}

	lAddr := config.P2P.ExternalAddress
	if lAddr == "" {
		lAddr = config.P2P.ListenAddress
	}
	addr, err := p2p.NewNetAddressFromString(p2p.NetAddressString(nodeKey.ID(), lAddr))
	if err != nil {
		return nodeInfo, errors.Wrap(err, "invalid (local) node net address")
	}
	nodeInfo.NetAddress = addr

	err = nodeInfo.Validate()
	return nodeInfo, err
}

// ------------------------------------------------------------------------------

var genesisDocKey = []byte("genesisDoc")

// LoadStateFromDBOrGenesisDocProvider attempts to load the state from the
// database, or creates one using the given genesisDocProvider and persists the
// result to the database. On success this also returns the genesis doc loaded
// through the given provider.
func LoadStateFromDBOrGenesisDocProvider(stateDB dbm.DB, genesisDocProvider GenesisDocProvider) (sm.State, *types.GenesisDoc, error) {
	// Get genesis doc
	genDoc, err := loadGenesisDoc(stateDB)
	if err != nil {
		genDoc, err = genesisDocProvider()
		if err != nil {
			return sm.State{}, nil, err
		}
		// save genesis doc to prevent a certain class of user errors (e.g. when it
		// was changed, accidentally or not). Also good for audit trail.
		saveGenesisDoc(stateDB, genDoc)
	}
	state, err := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
	if err != nil {
		return sm.State{}, nil, err
	}
	return state, genDoc, nil
}

// panics if failed to unmarshal bytes
func loadGenesisDoc(db dbm.DB) (*types.GenesisDoc, error) {
	b := db.Get(genesisDocKey)
	if len(b) == 0 {
		return nil, errors.New("Genesis doc not found")
	}
	var genDoc *types.GenesisDoc
	err := amino.UnmarshalJSON(b, &genDoc)
	if err != nil {
		panic(fmt.Sprintf("Failed to load genesis doc due to unmarshaling error: %v (bytes: %X)", err, b))
	}
	return genDoc, nil
}

// panics if failed to marshal the given genesis document
func saveGenesisDoc(db dbm.DB, genDoc *types.GenesisDoc) {
	b, err := amino.MarshalJSON(genDoc)
	if err != nil {
		panic(fmt.Sprintf("Failed to save genesis doc due to marshaling error: %v", err))
	}
	db.SetSync(genesisDocKey, b)
}

func createAndStartPrivValidatorSocketClient(
	listenAddr string,
	logger *slog.Logger,
) (types.PrivValidator, error) {
	pve, err := privval.NewSignerListener(listenAddr, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start private validator")
	}

	pvsc, err := privval.NewSignerClient(pve)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start private validator")
	}

	return pvsc, nil
}

// splitAndTrimEmpty slices s into all subslices separated by sep and returns a
// slice of the string s with all leading and trailing Unicode code points
// contained in cutset removed. If sep is empty, SplitAndTrim splits after each
// UTF-8 sequence. First part is equivalent to strings.SplitN with a count of
// -1.  also filter out empty strings, only return non-empty strings.
func splitAndTrimEmpty(s, sep, cutset string) []string {
	if s == "" {
		return []string{}
	}

	spl := strings.Split(s, sep)
	nonEmptyStrings := make([]string, 0, len(spl))
	for i := 0; i < len(spl); i++ {
		element := strings.Trim(spl[i], cutset)
		if element != "" {
			nonEmptyStrings = append(nonEmptyStrings, element)
		}
	}
	return nonEmptyStrings
}
