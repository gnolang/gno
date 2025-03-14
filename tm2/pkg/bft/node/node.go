package node

// Ignore pprof import for gosec, as profiling
// is enabled by the user by setting a profiling address

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/cors"

	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/backup"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/file"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/discovery"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/rs/cors"

	"github.com/gnolang/gno/tm2/pkg/amino"
	bc "github.com/gnolang/gno/tm2/pkg/bft/blockchain"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	cs "github.com/gnolang/gno/tm2/pkg/bft/consensus"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
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

// Reactors are hooks for the p2p module,
// to alert of connecting / disconnecting peers
const (
	mempoolReactorName    = "MEMPOOL"
	blockchainReactorName = "BLOCKCHAIN"
	consensusReactorName  = "CONSENSUS"
	discoveryReactorName  = "DISCOVERY"
)

const (
	mempoolModuleName    = "mempool"
	blockchainModuleName = "blockchain"
	consensusModuleName  = "consensus"
	p2pModuleName        = "p2p"
	discoveryModuleName  = "discovery"
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
func DefaultNewNode(
	config *cfg.Config,
	genesisFile string,
	evsw events.EventSwitch,
	logger *slog.Logger,
) (*Node, error) {
	// Generate node PrivKey
	nodeKey, err := p2pTypes.LoadOrMakeNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	// Get app client creator.
	appClientCreator := proxy.DefaultClientCreator(
		config.LocalApp,
		config.ProxyApp,
		config.ABCI,
		config.DBDir(),
	)

	// Initialize the privValidator
	privVal, err := privval.NewPrivValidatorFromConfig(
		config.Consensus.PrivValidator,
		nodeKey.PrivKey,
		logger.With("module", "remote_signer_client"),
	)
	if err != nil {
		return nil, err
	}

	return NewNode(
		config,
		privVal,
		nodeKey,
		appClientCreator,
		DefaultGenesisDocProviderFunc(genesisFile),
		DefaultDBProvider,
		evsw,
		logger,
	)
}

// Option sets a parameter for the node.
type Option func(*Node)

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
	transport        *p2p.MultiplexTransport
	sw               *p2p.MultiplexSwitch // p2p connections
	discoveryReactor *discovery.Reactor   // discovery reactor
	nodeInfo         p2pTypes.NodeInfo
	nodeKey          *p2pTypes.NodeKey // our node privkey
	isListening      bool

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
	backupServer      *http.Server
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
	return privVal.PubKey().Address() == addr
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
	mempoolLogger := logger.With("module", mempoolModuleName)
	mempoolReactor := mempl.NewReactor(config.Mempool, mempool)
	mempoolReactor.SetLogger(mempoolLogger)

	if config.Consensus.WaitForTxs() {
		mempool.EnableTxsAvailable()
	}
	return mempoolReactor, mempool
}

func createBlockchainReactor(
	state sm.State,
	blockExec *sm.BlockExecutor,
	blockStore *store.BlockStore,
	fastSync bool,
	switchToConsensusFn bc.SwitchToConsensusFn,
	logger *slog.Logger,
) (bcReactor p2p.Reactor, err error) {
	bcReactor = bc.NewBlockchainReactor(
		state.Copy(),
		blockExec,
		blockStore,
		fastSync,
		switchToConsensusFn,
	)

	bcReactor.SetLogger(logger.With("module", blockchainModuleName))
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

type nodeReactor struct {
	name    string
	reactor p2p.Reactor
}

// NewNode returns a new, ready to go, Tendermint Node.
func NewNode(config *cfg.Config,
	privValidator types.PrivValidator,
	nodeKey *p2pTypes.NodeKey,
	clientCreator appconn.ClientCreator,
	genesisDocProvider GenesisDocProvider,
	dbProvider DBProvider,
	evsw events.EventSwitch,
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

	// Signal readiness when the node produces or receives its first block.
	const readinessListenerID = "first_block_listener"

	cFirstBlock := make(chan struct{})
	if blockStore.Height() > 0 {
		close(cFirstBlock)
	} else {
		var once sync.Once
		evsw.AddListener(readinessListenerID, func(ev events.Event) {
			if _, ok := ev.(types.EventNewBlock); ok {
				once.Do(func() {
					close(cFirstBlock)
					evsw.RemoveListener(readinessListenerID)
				})
			}
		})
	}

	// Transaction event storing
	eventStoreService, txEventStore, err := createAndStartEventStoreService(config, evsw, logger)
	if err != nil {
		return nil, err
	}

	// Create the handshaker, which calls RequestInfo, sets the AppVersion on the state,
	// and replays any blocks as necessary to sync tendermint with the app.
	consensusLogger := logger.With("module", consensusModuleName)
	if err := doHandshake(stateDB, state, blockStore, genDoc, evsw, proxyApp, consensusLogger); err != nil {
		return nil, err
	}

	// Reload the state. It will have the Version.Consensus.App set by the
	// Handshake, and may have other modifications as well (ie. depending on
	// what happened during block replay).
	state = sm.LoadState(stateDB)

	logNodeStartupInfo(state, privValidator.PubKey(), logger, consensusLogger)

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

	// Make ConsensusReactor
	consensusReactor, consensusState := createConsensusReactor(
		config, state, blockExec, blockStore, mempool,
		privValidator, fastSync, evsw, consensusLogger,
	)

	// Make BlockchainReactor
	bcReactor, err := createBlockchainReactor(
		state,
		blockExec,
		blockStore,
		fastSync,
		consensusReactor.SwitchToConsensus,
		logger,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create blockchain reactor")
	}

	reactors := []nodeReactor{
		{
			mempoolReactorName, mempoolReactor,
		},
		{
			blockchainReactorName, bcReactor,
		},
		{
			consensusReactorName, consensusReactor,
		},
	}

	nodeInfo, err := makeNodeInfo(config, nodeKey, txEventStore, genDoc, state)
	if err != nil {
		return nil, errors.Wrap(err, "error making NodeInfo")
	}

	p2pLogger := logger.With("module", p2pModuleName)

	// Setup the multiplex transport, used by the P2P switch
	transport := p2p.NewMultiplexTransport(
		nodeInfo,
		*nodeKey,
		conn.MConfigFromP2P(config.P2P),
		p2pLogger.With("transport", "multiplex"),
	)

	var discoveryReactor *discovery.Reactor

	if config.P2P.PeerExchange {
		discoveryReactor = discovery.NewReactor()

		discoveryReactor.SetLogger(logger.With("module", discoveryModuleName))

		reactors = append(reactors, nodeReactor{
			name:    discoveryReactorName,
			reactor: discoveryReactor,
		})
	}

	// Setup MultiplexSwitch.
	peerAddrs, errs := p2pTypes.NewNetAddressFromStrings(
		splitAndTrimEmpty(config.P2P.PersistentPeers, ",", " "),
	)
	for _, err = range errs {
		p2pLogger.Error("invalid persistent peer address", "err", err)
	}

	// Parse the private peer IDs
	privatePeerIDs, errs := p2pTypes.NewIDFromStrings(
		splitAndTrimEmpty(config.P2P.PrivatePeerIDs, ",", " "),
	)
	for _, err = range errs {
		p2pLogger.Error("invalid private peer ID", "err", err)
	}

	// Prepare the misc switch options
	opts := []p2p.SwitchOption{
		p2p.WithPersistentPeers(peerAddrs),
		p2p.WithPrivatePeers(privatePeerIDs),
		p2p.WithMaxInboundPeers(config.P2P.MaxNumInboundPeers),
		p2p.WithMaxOutboundPeers(config.P2P.MaxNumOutboundPeers),
	}

	// Prepare the reactor switch options
	for _, r := range reactors {
		opts = append(opts, p2p.WithReactor(r.name, r.reactor))
	}

	sw := p2p.NewMultiplexSwitch(
		transport,
		opts...,
	)

	sw.SetLogger(p2pLogger)

	p2pLogger.Info("P2P Node ID", "ID", nodeKey.ID(), "file", config.NodeKeyFile())

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

		transport:        transport,
		sw:               sw,
		discoveryReactor: discoveryReactor,
		nodeInfo:         nodeInfo,
		nodeKey:          nodeKey,

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

	// start backup server if requested
	if n.config.Backup != nil && n.config.Backup.ListenAddress != "" {
		n.backupServer = backup.NewServer(n.config.Backup, n.blockStore)
		go func() {
			if err := n.backupServer.ListenAndServe(); err != nil {
				n.Logger.Error("Backup server", "err", err)
			}
		}()
	}

	// Start the transport.
	// The listen address for the transport needs to be an address within reach of the machine NIC
	listenAddress := p2pTypes.NetAddressString(n.nodeKey.ID(), n.config.P2P.ListenAddress)

	addr, err := p2pTypes.NewNetAddressFromString(listenAddress)
	if err != nil {
		return fmt.Errorf("unable to parse network address, %w", err)
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
	peerAddrs, errs := p2pTypes.NewNetAddressFromStrings(splitAndTrimEmpty(n.config.P2P.PersistentPeers, ",", " "))
	for _, err := range errs {
		n.Logger.Error("invalid persistent peer address", "err", err)
	}

	// Dial the persistent peers
	n.sw.DialPeers(peerAddrs...)

	return nil
}

// OnStop stops the Node. It implements service.Service.
func (n *Node) OnStop() {
	n.BaseService.OnStop()

	n.Logger.Info("Stopping Node")

	// Fist close the private validator
	if err := n.privValidator.Close(); err != nil {
		n.Logger.Error("Error closing private validator", "err", err)
	}

	// Stop the non-reactor services
	n.evsw.Stop()
	n.eventStoreService.Stop()

	// Stop the node p2p transport
	if err := n.transport.Close(); err != nil {
		n.Logger.Error("unable to gracefully close transport", "err", err)
	}

	// now stop the reactors
	if err := n.sw.Stop(); err != nil {
		n.Logger.Error("unable to gracefully close switch", "err", err)
	}

	// stop mempool WAL
	if n.config.Mempool.WalEnabled() {
		n.mempool.CloseWAL()
	}

	n.isListening = false

	// stop the backup server if started
	if n.backupServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := n.backupServer.Shutdown(ctx); err != nil {
			n.Logger.Error("Error closing backup server", "err", err)
		}
	}

	// finally stop the listeners / external services
	for _, l := range n.rpcListeners {
		n.Logger.Info("Closing rpc listener", "listener", l)
		if err := l.Close(); err != nil {
			n.Logger.Error("Error closing listener", "listener", l, "err", err)
		}
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
	rpccore.SetPubKey(n.privValidator.PubKey())
	rpccore.SetGenesisDoc(n.genesisDoc)
	rpccore.SetProxyAppQuery(n.proxyApp.Query())
	rpccore.SetGetFastSync(n.consensusReactor.FastSync)
	rpccore.SetLogger(n.Logger.With("module", "rpc"))
	rpccore.SetEventSwitch(n.evsw)
	rpccore.SetConfig(*n.config.RPC)
}

func (n *Node) startRPC() (listeners []net.Listener, err error) {
	defer func() {
		if err != nil {
			// Close all the created listeners on any error, instead of
			// leaking them: https://github.com/gnolang/gno/issues/3639
			for _, ln := range listeners {
				ln.Close()
			}
		}
	}()

	listenAddrs := splitAndTrimEmpty(n.config.RPC.ListenAddress, ",", " ")

	config := rpcserver.DefaultConfig()
	config.MaxBodyBytes = n.config.RPC.MaxBodyBytes
	config.MaxHeaderBytes = n.config.RPC.MaxHeaderBytes
	config.MaxOpenConnections = n.config.RPC.MaxOpenConnections
	// If necessary adjust global WriteTimeout to ensure it's greater than
	// TimeoutBroadcastTxCommit.
	// See https://github.com/tendermint/tendermint/issues/3435
	if config.WriteTimeout <= n.config.RPC.TimeoutBroadcastTxCommit {
		config.WriteTimeout = n.config.RPC.TimeoutBroadcastTxCommit + 1*time.Second
	}

	// we may expose the rpc over both a unix and tcp socket
	var rebuildAddresses bool
	listeners = make([]net.Listener, 0, len(listenAddrs))
	for _, listenAddr := range listenAddrs {
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

		listeners = append(listeners, listener)
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
func (n *Node) Switch() *p2p.MultiplexSwitch {
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
func (n *Node) NodeInfo() p2pTypes.NodeInfo {
	return n.nodeInfo
}

func makeNodeInfo(
	config *cfg.Config,
	nodeKey *p2pTypes.NodeKey,
	txEventStore eventstore.TxEventStore,
	genDoc *types.GenesisDoc,
	state sm.State,
) (p2pTypes.NodeInfo, error) {
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

	nodeInfo := p2pTypes.NodeInfo{
		VersionSet: vset,
		NetAddress: nil, // The shared address depends on the configuration
		Network:    genDoc.ChainID,
		Version:    version.Version,
		Channels: []byte{
			bcChannel,
			cs.StateChannel, cs.DataChannel, cs.VoteChannel, cs.VoteSetBitsChannel,
			mempl.MempoolChannel,
		},
		Moniker: config.Moniker,
		Other: p2pTypes.NodeInfoOther{
			TxIndex:    txIndexerStatus,
			RPCAddress: config.RPC.ListenAddress,
		},
	}

	// Make sure the discovery channel is shared with peers
	// in case peer discovery is enabled
	if config.P2P.PeerExchange {
		nodeInfo.Channels = append(nodeInfo.Channels, discovery.Channel)
	}

	// Grab the supplied listen address.
	// This address needs to be valid, but it can be unspecified.
	// If the listen address is unspecified (port / IP unbound),
	// then this address cannot be used by peers for dialing
	addr, err := p2pTypes.NewNetAddressFromString(
		p2pTypes.NetAddressString(nodeKey.ID(), config.P2P.ListenAddress),
	)
	if err != nil {
		return p2pTypes.NodeInfo{}, fmt.Errorf("unable to parse network address, %w", err)
	}

	// Use the transport listen address as the advertised address
	nodeInfo.NetAddress = addr

	// Prepare the advertised dial address (if any)
	// for the node, which other peers can use to dial
	if config.P2P.ExternalAddress != "" {
		addr, err = p2pTypes.NewNetAddressFromString(
			p2pTypes.NetAddressString(
				nodeKey.ID(),
				config.P2P.ExternalAddress,
			),
		)
		if err != nil {
			return p2pTypes.NodeInfo{}, fmt.Errorf("invalid p2p external address: %w", err)
		}

		nodeInfo.NetAddress = addr
	}

	// Validate the node info
	if err := nodeInfo.Validate(); err != nil {
		return p2pTypes.NodeInfo{}, fmt.Errorf("unable to validate node info, %w", err)
	}

	return nodeInfo, nil
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
	b, err := db.Get(genesisDocKey)
	if err != nil {
		return nil, fmt.Errorf("error while getting Genesis doc: %w", err)
	}
	if len(b) == 0 {
		return nil, errors.New("Genesis doc not found")
	}
	var genDoc *types.GenesisDoc
	err = amino.UnmarshalJSON(b, &genDoc)
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
	for i := range spl {
		element := strings.Trim(spl[i], cutset)
		if element != "" {
			nonEmptyStrings = append(nonEmptyStrings, element)
		}
	}
	return nonEmptyStrings
}
