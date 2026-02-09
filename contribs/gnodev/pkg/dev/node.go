package dev

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tm2events "github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type NodeConfig struct {
	// Logger is used for logging node activities. It can be set to a custom logger or a noop logger for
	// silent operation.
	Logger *slog.Logger

	// Loader is responsible for loading packages. It abstracts the mechanism for retrieving and managing
	// package data.
	Loader packages.Loader

	// DefaultCreator specifies the default address used for creating packages and transactions.
	DefaultCreator crypto.Address

	// DefaultDeposit is the default amount of coins deposited when creating a package.
	DefaultDeposit std.Coins

	// BalancesList defines the initial balance of accounts in the genesis state.
	BalancesList []gnoland.Balance

	// PackagesModifier allows modifications to be applied to packages during initialization.
	PackagesModifier []QueryPath

	// Emitter is used to emit events for various node operations. It can be set to a noop emitter if no
	// event emission is required.
	Emitter emitter.Emitter

	// InitialTxs contains the transactions that are included in the genesis state.
	InitialTxs []gnoland.TxWithMetadata

	// TMConfig holds the Tendermint configuration settings.
	TMConfig *tmcfg.Config

	// SkipFailingGenesisTxs indicates whether to skip failing transactions during the genesis
	// initialization.
	SkipFailingGenesisTxs bool

	// NoReplay, if set to true, prevents replaying of transactions from the block store during node
	// initialization.
	NoReplay bool

	// MaxGasPerBlock sets the maximum amount of gas that can be used in a single block.
	MaxGasPerBlock int64

	// ChainID is the unique identifier for the blockchain.
	ChainID string

	// ChainDomain specifies the domain name associated with the blockchain network.
	ChainDomain string
}

func DefaultNodeConfig(rootdir, domain string) *NodeConfig {
	tmc := gnoland.NewDefaultTMConfig(rootdir)
	tmc.Consensus.SkipTimeoutCommit = false // avoid time drifting, see issue #1507
	tmc.Consensus.WALDisabled = true
	tmc.Consensus.CreateEmptyBlocks = false

	defaultDeployer := crypto.MustAddressFromString(integration.DefaultAccount_Address)
	balances := []gnoland.Balance{
		{
			Address: defaultDeployer,
			Amount:  std.Coins{std.NewCoin(ugnot.Denom, 10e12)},
		},
	}

	exampleFolder := filepath.Join(gnoenv.RootDir(), "example") // XXX: we should avoid having to hardcoding this here
	defaultLoader := packages.NewLoader(packages.NewRootResolver(exampleFolder))

	return &NodeConfig{
		Logger:                log.NewNoopLogger(),
		Emitter:               &emitter.NoopServer{},
		Loader:                defaultLoader,
		DefaultCreator:        defaultDeployer,
		DefaultDeposit:        nil,
		BalancesList:          balances,
		ChainID:               tmc.ChainID(),
		ChainDomain:           domain,
		TMConfig:              tmc,
		SkipFailingGenesisTxs: true,
		MaxGasPerBlock:        10_000_000_000,
	}
}

// Node is not thread safe
type Node struct {
	*node.Node
	muNode sync.RWMutex

	config       *NodeConfig
	emitter      emitter.Emitter
	client       client.Client
	logger       *slog.Logger
	loader       packages.Loader
	pkgs         []packages.Package
	pkgsModifier map[string]QueryPath // path -> QueryPath
	paths        []string

	// keep track of number of loaded package to be able to skip them on restore
	loadedPackages int

	// track starting time for genesis
	startTime time.Time

	// state
	initialState, state []gnoland.TxWithMetadata
	currentStateIndex   int
}

var DefaultFee = std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))

func NewDevNode(ctx context.Context, cfg *NodeConfig, pkgpaths ...string) (*Node, error) {
	startTime := time.Now()

	pkgsModifier := make(map[string]QueryPath, len(cfg.PackagesModifier))
	for _, qpath := range cfg.PackagesModifier {
		pkgsModifier[qpath.Path] = qpath
	}

	devnode := &Node{
		loader:            cfg.Loader,
		config:            cfg,
		emitter:           cfg.Emitter,
		logger:            cfg.Logger,
		startTime:         startTime,
		state:             cfg.InitialTxs,
		initialState:      cfg.InitialTxs,
		currentStateIndex: len(cfg.InitialTxs),
		paths:             pkgpaths,
		pkgsModifier:      pkgsModifier,
	}

	// XXX: MOVE THIS, passing context here can be confusing
	if err := devnode.Reset(ctx); err != nil {
		return nil, fmt.Errorf("unable to initialize the node: %w", err)
	}

	return devnode, nil
}

func (n *Node) Paths() []string {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	return n.paths
}

func (n *Node) Close() error {
	n.muNode.Lock()
	defer n.muNode.Unlock()

	return n.Node.Stop()
}

func (n *Node) ListPkgs() []packages.Package {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	return n.pkgs
}

func (n *Node) Client() client.Client {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	return n.client
}

func (n *Node) GetRemoteAddress() string {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	return n.Node.RPC().ListenAddress()
}

// AddPackagePaths to load
func (n *Node) AddPackagePaths(paths ...string) {
	n.muNode.Lock()
	defer n.muNode.Unlock()

	n.paths = append(n.paths, paths...)
}

func (n *Node) SetPackagePaths(paths ...string) {
	n.muNode.Lock()
	defer n.muNode.Unlock()

	n.paths = paths
}

// HasPackageLoaded returns true if the specified package has already been loaded.
// NOTE: This only checks if the package was loaded at the genesis level.
func (n *Node) HasPackageLoaded(path string) bool {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	for _, pkg := range n.pkgs {
		if pkg.MemPackage.Path == path {
			return true
		}
	}

	return false
}

// GetBlockTransactions returns the transactions contained
// within the specified block, if any.
func (n *Node) GetBlockTransactions(ctx context.Context, blockNum uint64) ([]gnoland.TxWithMetadata, error) {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	return n.getBlockTransactions(ctx, blockNum)
}

// GetBlockTransactions returns the transactions contained
// within the specified block, if any.
func (n *Node) getBlockTransactions(ctx context.Context, blockNum uint64) ([]gnoland.TxWithMetadata, error) {
	int64BlockNum := int64(blockNum)
	b, err := n.client.Block(ctx, &int64BlockNum)
	if err != nil {
		return nil, fmt.Errorf("unable to load block at height %d: %w", blockNum, err)
	}
	txs := b.Block.Data.Txs

	bres, err := n.client.BlockResults(ctx, &int64BlockNum)
	if err != nil {
		return nil, fmt.Errorf("unable to load block at height %d: %w", blockNum, err)
	}
	deliverTxs := bres.Results.DeliverTxs

	// Sanity check
	if len(txs) != len(deliverTxs) {
		panic(fmt.Errorf("invalid block txs len (%d) vs block result txs len (%d)",
			len(txs), len(deliverTxs),
		))
	}

	txResults := make([]*abci.ResponseDeliverTx, len(deliverTxs))
	for i, tx := range deliverTxs {
		txResults[i] = &tx
	}

	// XXX: Consider replacing a failed transaction with an empty transaction
	// to preserve the transaction height ?
	// Note that this would also require committing instead of using the
	// genesis block.

	metaTxs := make([]gnoland.TxWithMetadata, 0, len(txs))
	for i, encodedTx := range txs {
		if deliverTx := deliverTxs[i]; !deliverTx.IsOK() {
			continue // skip failed tx
		}

		var tx std.Tx
		if unmarshalErr := amino.Unmarshal(encodedTx, &tx); unmarshalErr != nil {
			return nil, fmt.Errorf("unable to unmarshal tx: %w", unmarshalErr)
		}

		metaTxs = append(metaTxs, gnoland.TxWithMetadata{
			Tx: tx,
			Metadata: &gnoland.GnoTxMetadata{
				Timestamp: b.BlockMeta.Header.Time.Unix(),
			},
		})
	}

	return slices.Clip(metaTxs), nil
}

// GetBlockTransactions returns the transactions contained
// within the specified block, if any.
// GetLatestBlockNumber returns the latest block height from the chain.
func (n *Node) GetLatestBlockNumber() (uint64, error) {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	return n.getLatestBlockNumber(), nil
}

func (n *Node) getLatestBlockNumber() uint64 {
	return uint64(n.Node.BlockStore().Height())
}

// Reset stops the node, if running, and reloads it with a new genesis state,
// effectively ignoring the current state.
func (n *Node) Reset(ctx context.Context) error {
	n.muNode.Lock()
	defer n.muNode.Unlock()

	// Reset starting time
	startTime := time.Now()

	// Generate a new genesis state based on the current packages
	pkgs, err := n.loader.Load(n.paths...)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	// Append initialTxs
	pkgsTxs := n.generateTxs(DefaultFee, pkgs)
	txs := append(pkgsTxs, n.initialState...)

	genesis := gnoland.DefaultGenState()
	genesis.Balances = n.config.BalancesList
	genesis.Txs = txs

	// Reset the node with the new genesis state.
	err = n.rebuildNode(ctx, genesis)
	if err != nil {
		return fmt.Errorf("unable to initialize a new node: %w", err)
	}

	n.pkgs = pkgs
	n.loadedPackages = len(pkgsTxs)
	n.currentStateIndex = len(n.initialState)
	n.startTime = startTime
	n.emitter.Emit(&events.Reset{})
	return nil
}

// ReloadAll updates all currently known packages and then reloads the node.
// It's actually a simple combination between `UpdatePackage` and `Reload` method.
func (n *Node) ReloadAll(ctx context.Context) error {
	n.muNode.Lock()
	defer n.muNode.Unlock()

	return n.rebuildNodeFromState(ctx)
}

// Reload saves the current state, stops the node if running, starts a new node,
// and re-apply previously saved state along with packages updated by `UpdatePackages`.
// If any transaction, including 'addpkg', fails, it will be ignored.
// Use 'Reset' to completely reset the node's state in case of persistent errors.
func (n *Node) Reload(ctx context.Context) error {
	n.muNode.Lock()
	defer n.muNode.Unlock()

	return n.rebuildNodeFromState(ctx)
}

// SendTransaction executes a broadcast commit send
// of the specified transaction to the chain
func (n *Node) SendTransaction(tx *std.Tx) error {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	aminoTx, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal transaction to amino binary, %w", err)
	}

	// we use BroadcastTxCommit to ensure to have one block with the given tx
	res, err := n.client.BroadcastTxCommit(context.Background(), aminoTx)
	if err != nil {
		return fmt.Errorf("unable to broadcast transaction commit: %w", err)
	}

	if res.CheckTx.Error != nil {
		n.logger.Error("check tx error trace", "log", res.CheckTx.Log)
		return fmt.Errorf("check transaction error: %w", res.CheckTx.Error)
	}

	if res.DeliverTx.Error != nil {
		n.logger.Error("deliver tx error trace", "log", res.CheckTx.Log)
		return fmt.Errorf("deliver transaction error: %w", res.DeliverTx.Error)
	}

	return nil
}

func (n *Node) getBlockStoreState(ctx context.Context) ([]gnoland.TxWithMetadata, error) {
	// get current genesis state
	genesis := n.GenesisDoc().AppState.(gnoland.GnoGenesisState)

	initialTxs := genesis.Txs[n.loadedPackages:] // ignore previously loaded packages
	state := append([]gnoland.TxWithMetadata{}, initialTxs...)

	lastBlock := n.getLatestBlockNumber()
	var blocnum uint64 = 1
	for ; blocnum <= lastBlock; blocnum++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		txs, txErr := n.getBlockTransactions(ctx, blocnum)
		if txErr != nil {
			return nil, fmt.Errorf("unable to fetch block transactions, %w", txErr)
		}

		state = append(state, txs...)
	}

	return state, nil
}

func (n *Node) generateTxs(fee std.Fee, pkgs []packages.Package) []gnoland.TxWithMetadata {
	metatxs := make([]gnoland.TxWithMetadata, 0, len(pkgs))
	for _, pkg := range pkgs {
		msg := vm.MsgAddPackage{
			Creator:    n.config.DefaultCreator,
			MaxDeposit: n.config.DefaultDeposit,
			Package:    &pkg.MemPackage,
		}

		if m, ok := n.pkgsModifier[pkg.Path]; ok {
			if !m.Creator.IsZero() {
				msg.Creator = m.Creator
			}

			if m.Deposit != nil {
				msg.MaxDeposit = m.Deposit
			}

			n.logger.Debug("applying pkgs modifier",
				"path", pkg.Path,
				"creator", msg.Creator,
				"deposit", msg.MaxDeposit,
			)
		}

		// Create transaction
		tx := std.Tx{Fee: fee, Msgs: []std.Msg{msg}}
		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))

		// Wrap it with metadata
		metatx := gnoland.TxWithMetadata{
			Tx: tx,
			Metadata: &gnoland.GnoTxMetadata{
				Timestamp: n.startTime.Unix(),
			},
		}
		metatxs = append(metatxs, metatx)
	}

	return metatxs
}

func (n *Node) stopIfRunning() error {
	if n.Node != nil && n.Node.IsRunning() {
		if err := n.Node.Stop(); err != nil {
			return fmt.Errorf("unable to stop the node: %w", err)
		}
	}

	return nil
}

func (n *Node) rebuildNodeFromState(ctx context.Context) error {
	start := time.Now()

	if n.config.NoReplay {
		// If NoReplay is true, simply reset the node to its initial state
		n.logger.Warn("replay disabled")

		pkgs, err := n.loader.Load(n.paths...)
		if err != nil {
			return fmt.Errorf("unable to load pkgs: %w", err)
		}

		genesis := gnoland.DefaultGenState()
		genesis.Balances = n.config.BalancesList
		genesis.Txs = n.generateTxs(DefaultFee, pkgs)
		return n.rebuildNode(ctx, genesis)
	}

	state, err := n.getBlockStoreState(ctx)
	if err != nil {
		return fmt.Errorf("unable to save state: %s", err.Error())
	}

	// Load genesis packages
	pkgs, err := n.loader.Load(n.paths...)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	// Create genesis with loaded pkgs + previous state
	genesis := gnoland.DefaultGenState()
	genesis.Balances = n.config.BalancesList

	// Generate txs
	pkgsTxs := n.generateTxs(DefaultFee, pkgs)
	genesis.Txs = append(pkgsTxs, state...)

	// Reset the node with the new genesis state.
	err = n.rebuildNode(ctx, genesis)
	if err != nil {
		return fmt.Errorf("unable to rebuild node: %w", err)
	}
	n.logger.Info("reload done",
		"pkgs", len(pkgsTxs),
		"state applied", len(state),
		"took", time.Since(start),
	)

	// Update node infos
	n.pkgs = pkgs
	n.loadedPackages = len(pkgsTxs)

	// Emit reload event
	n.emitter.Emit(&events.Reload{})
	return nil
}

func (n *Node) handleEventTX(evt tm2events.Event) {
	switch data := evt.(type) {
	case bft.EventTx:
		go func() {
			// Use a separate goroutine in order to avoid a deadlock situation.
			// This is needed because this callback may get called during node rebuilding while
			// lock is held.
			n.muNode.Lock()
			defer n.muNode.Unlock()

			heigh := n.BlockStore().Height()
			n.currentStateIndex++
			n.state = nil // invalidate state

			n.logger.Info("node state", "index", n.currentStateIndex, "height", heigh)
		}()

		resEvt := events.TxResult{
			Height: data.Result.Height,
			Index:  data.Result.Index,
			// XXX: Update this to split error for stack
			Response: data.Result.Response,
		}

		if err := amino.Unmarshal(data.Result.Tx, &resEvt.Tx); err != nil {
			n.logger.Error("unable to unwrap tx result",
				"error", err)
		}

		n.emitter.Emit(resEvt)
	}
}

func (n *Node) rebuildNode(ctx context.Context, genesis gnoland.GnoGenesisState) (err error) {
	noopLogger := log.NewNoopLogger()

	// Stop the node if it's currently running.
	if err := n.stopIfRunning(); err != nil {
		return fmt.Errorf("unable to stop the node: %w", err)
	}

	// Setup node config
	nodeConfig := newNodeConfig(n.config.TMConfig, n.config.ChainID, n.config.ChainDomain, genesis)
	nodeConfig.GenesisTxResultHandler = n.genesisTxResultHandler
	// Speed up stdlib loading after first start (saves about 2-3 seconds on each reload).
	nodeConfig.CacheStdlibLoad = true
	nodeConfig.Genesis.ConsensusParams.Block.MaxGas = n.config.MaxGasPerBlock
	// Genesis verification is always false with Gnodev
	nodeConfig.SkipGenesisSigVerification = true

	// recoverFromError handles panics and converts them to errors.
	recoverFromError := func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case error:
				err = val
			case string:
				err = fmt.Errorf("error: %s", val)
			default:
				err = fmt.Errorf("unknown error: %#v", val)
			}
		}
	}

	// Execute node creation and handle any errors.
	defer recoverFromError()

	// XXX: Redirect the node log somewhere else
	node, nodeErr := gnoland.NewInMemoryNode(noopLogger, nodeConfig)
	if nodeErr != nil {
		return fmt.Errorf("unable to create a new node: %w", err)
	}

	node.EventSwitch().AddListener("dev-emitter", n.handleEventTX)

	if startErr := node.Start(); startErr != nil {
		return fmt.Errorf("unable to start the node: %w", startErr)
	}

	// Wait for the node to be ready
	select {
	case <-node.Ready(): // Ok
		n.Node = node
	case <-ctx.Done():
		return ctx.Err()
	}

	// Create the RPC client using the actual bound address.
	// This is done after the node starts because the listen address
	// may use port 0, and the actual port is only known after binding
	rpcClient, err := client.NewHTTPClient(node.RPC().ListenAddress())
	if err != nil {
		return fmt.Errorf("unable to create RPC client: %w", err)
	}
	n.client = rpcClient

	return nil
}

func (n *Node) genesisTxResultHandler(ctx sdk.Context, tx std.Tx, res sdk.Result) {
	if !res.IsErr() {
		for _, msg := range tx.Msgs {
			if addpkg, ok := msg.(vm.MsgAddPackage); ok && addpkg.Package != nil {
				n.logger.Debug("add package",
					"path", addpkg.Package.Path,
					"files", len(addpkg.Package.Files),
					"creator", addpkg.Creator.String(),
				)
			}
		}

		return
	}

	// XXX: for now, this is only way to catch the error
	before, after, found := strings.Cut(res.Log, "\n")
	if !found {
		n.logger.Error("unable to send tx", "log", res.Log)
		return
	}

	var attrs []slog.Attr

	// Add error
	attrs = append(attrs, slog.Any("err", res.Error))

	// Fetch first line as error message
	msg := strings.TrimFunc(before, func(r rune) bool {
		return unicode.IsSpace(r) || r == ':'
	})
	attrs = append(attrs, slog.String("err", msg))

	// If debug is enable, also append stack
	if n.logger.Enabled(context.Background(), slog.LevelDebug) {
		attrs = append(attrs, slog.String("stack", after))
	}

	n.logger.LogAttrs(context.Background(), slog.LevelError, "unable to deliver tx", attrs...)
}

func newNodeConfig(tmc *tmcfg.Config, chainid, chaindomain string, appstate gnoland.GnoGenesisState) *gnoland.InMemoryNodeConfig {
	// Create Mocked Identity
	pv := bft.NewMockPV()
	genesis := gnoland.NewDefaultGenesisConfig(chainid, chaindomain)
	genesis.AppState = appstate

	// Add self as validator
	self := pv.PubKey()
	genesis.Validators = []bft.GenesisValidator{
		{
			Address: self.Address(),
			PubKey:  self,
			Power:   10,
			Name:    "self",
		},
	}

	cfg := &gnoland.InMemoryNodeConfig{
		PrivValidator: pv,
		TMConfig:      tmc,
		Genesis:       genesis,
		VMOutput:      os.Stdout,
	}
	return cfg
}
