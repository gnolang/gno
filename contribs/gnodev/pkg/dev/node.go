package dev

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/amino"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tm2events "github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/std"
	// backup "github.com/gnolang/tx-archive/backup/client"
	// restore "github.com/gnolang/tx-archive/restore/client"
)

type NodeConfig struct {
	PackagesPathList      []string
	TMConfig              *tmcfg.Config
	SkipFailingGenesisTxs bool
	NoReplay              bool
	MaxGasPerBlock        int64
	ChainID               string
}

func DefaultNodeConfig(rootdir string) *NodeConfig {
	tmc := gnoland.NewDefaultTMConfig(rootdir)
	tmc.Consensus.SkipTimeoutCommit = false // avoid time drifting, see issue #1507

	return &NodeConfig{
		ChainID:               tmc.ChainID(),
		PackagesPathList:      []string{},
		TMConfig:              tmc,
		SkipFailingGenesisTxs: true,
		MaxGasPerBlock:        10_000_000_000,
	}
}

// Node is not thread safe
type Node struct {
	*node.Node

	config  *NodeConfig
	emitter emitter.Emitter
	client  client.Client
	logger  *slog.Logger
	pkgs    PkgsMap // path -> pkg

	// keep track of number of loaded package to be able to skip them on restore
	loadedPackages int
}

var (
	DefaultFee     = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
	DefaultCreator = crypto.MustAddressFromString(integration.DefaultAccount_Address)
	DefaultBalance = []gnoland.Balance{
		{
			Address: DefaultCreator,
			Amount:  std.MustParseCoins("10000000000000ugnot"),
		},
	}
)

func NewDevNode(ctx context.Context, logger *slog.Logger, emitter emitter.Emitter, cfg *NodeConfig) (*Node, error) {
	mpkgs, err := newPkgsMap(cfg.PackagesPathList)
	if err != nil {
		return nil, fmt.Errorf("unable map pkgs list: %w", err)
	}

	pkgsTxs, err := mpkgs.Load(DefaultCreator, DefaultFee, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to load genesis packages: %w", err)
	}

	// generate genesis state
	genesis := gnoland.GnoGenesisState{
		Balances: DefaultBalance,
		Txs:      pkgsTxs,
	}

	devnode := &Node{
		config:         cfg,
		emitter:        emitter,
		client:         client.NewLocal(),
		pkgs:           mpkgs,
		logger:         logger,
		loadedPackages: len(pkgsTxs),
	}

	if err := devnode.reset(ctx, genesis); err != nil {
		return nil, fmt.Errorf("unable to initialize the node: %w", err)
	}

	return devnode, nil
}

func (d *Node) getLatestBlockNumber() uint64 {
	return uint64(d.Node.BlockStore().Height())
}

func (d *Node) Close() error {
	return d.Node.Stop()
}

func (d *Node) ListPkgs() []gnomod.Pkg {
	return d.pkgs.toList()
}

func (d *Node) GetNodeReadiness() <-chan struct{} {
	return gnoland.GetNodeReadiness(d.Node)
}

func (d *Node) GetRemoteAddress() string {
	return d.Node.Config().RPC.ListenAddress
}

// UpdatePackages updates the currently known packages. It will be taken into
// consideration in the next reload of the node.
func (d *Node) UpdatePackages(paths ...string) error {
	for _, path := range paths {
		// List all packages from target path
		pkgslist, err := gnomod.ListPkgs(path)
		if err != nil {
			return fmt.Errorf("failed to list gno packages for %q: %w", path, err)
		}

		// Update or add package in the current known list.
		for _, pkg := range pkgslist {
			d.pkgs[pkg.Dir] = pkg
		}
	}

	return nil
}

// Reset stops the node, if running, and reloads it with a new genesis state,
// effectively ignoring the current state.
func (d *Node) Reset(ctx context.Context) error {
	// Stop the node if it's currently running.
	if err := d.stopIfRunning(); err != nil {
		return fmt.Errorf("unable to stop the node: %w", err)
	}

	// Generate a new genesis state based on the current packages
	txs, err := d.pkgs.Load(DefaultCreator, DefaultFee, nil)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	genesis := gnoland.GnoGenesisState{
		Balances: DefaultBalance,
		Txs:      txs,
	}

	// Reset the node with the new genesis state.
	err = d.reset(ctx, genesis)
	if err != nil {
		return fmt.Errorf("unable to initialize a new node: %w", err)
	}

	d.emitter.Emit(&events.Reset{})
	return nil
}

// ReloadAll updates all currently known packages and then reloads the node.
func (d *Node) ReloadAll(ctx context.Context) error {
	pkgs := d.ListPkgs()
	paths := make([]string, len(pkgs))
	for i, pkg := range pkgs {
		paths[i] = pkg.Dir
	}

	if err := d.UpdatePackages(paths...); err != nil {
		return fmt.Errorf("unable to reload packages: %w", err)
	}

	return d.Reload(ctx)
}

// Reload saves the current state, stops the node if running, starts a new node,
// and re-apply previously saved state along with packages updated by `UpdatePackages`.
// If any transaction, including 'addpkg', fails, it will be ignored.
// Use 'Reset' to completely reset the node's state in case of persistent errors.
func (d *Node) Reload(ctx context.Context) error {
	if d.config.NoReplay {
		// If NoReplay is true, reload as the same effect as reset
		d.logger.Warn("replay disable")
		return d.Reset(ctx)
	}

	// Get current blockstore state
	state, err := d.getBlockStoreState(ctx)
	if err != nil {
		return fmt.Errorf("unable to save state: %s", err.Error())
	}

	// Stop the node if it's currently running.
	if err := d.stopIfRunning(); err != nil {
		return fmt.Errorf("unable to stop the node: %w", err)
	}

	// Load genesis packages
	pkgsTxs, err := d.pkgs.Load(DefaultCreator, DefaultFee, nil)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	// Create genesis with loaded pkgs + previous state
	genesis := gnoland.GnoGenesisState{
		Balances: DefaultBalance,
		Txs:      append(pkgsTxs, state...),
	}

	// Reset the node with the new genesis state.
	err = d.reset(ctx, genesis)
	d.logger.Info("reload done", "pkgs", len(pkgsTxs), "state applied", len(state))

	// Update node infos
	d.loadedPackages = len(pkgsTxs)

	d.emitter.Emit(&events.Reload{})
	return nil
}

// GetBlockTransactions returns the transactions contained
// within the specified block, if any
func (d *Node) GetBlockTransactions(blockNum uint64) ([]std.Tx, error) {
	int64BlockNum := int64(blockNum)
	b, err := d.client.Block(&int64BlockNum)
	if err != nil {
		return []std.Tx{}, fmt.Errorf("unable to load block at height %d: %w", blockNum, err) // nothing to see here
	}

	txs := make([]std.Tx, len(b.Block.Data.Txs))
	for i, encodedTx := range b.Block.Data.Txs {
		var tx std.Tx
		if unmarshalErr := amino.Unmarshal(encodedTx, &tx); unmarshalErr != nil {
			return nil, fmt.Errorf("unable to unmarshal amino tx, %w", unmarshalErr)
		}

		txs[i] = tx
	}

	return txs, nil
}

// GetBlockTransactions returns the transactions contained
// within the specified block, if any
// GetLatestBlockNumber returns the latest block height from the chain
func (d *Node) GetLatestBlockNumber() (uint64, error) {
	return d.getLatestBlockNumber(), nil
}

// SendTransaction executes a broadcast commit send
// of the specified transaction to the chain
func (d *Node) SendTransaction(tx *std.Tx) error {
	aminoTx, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal transaction to amino binary, %w", err)
	}

	// we use BroadcastTxCommit to ensure to have one block with the given tx
	res, err := d.client.BroadcastTxCommit(aminoTx)
	if err != nil {
		return fmt.Errorf("unable to broadcast transaction commit: %w", err)
	}

	if res.CheckTx.Error != nil {
		d.logger.Error("check tx error trace", "log", res.CheckTx.Log)
		return fmt.Errorf("check transaction error: %w", res.CheckTx.Error)
	}

	if res.DeliverTx.Error != nil {
		d.logger.Error("deliver tx error trace", "log", res.CheckTx.Log)
		return fmt.Errorf("deliver transaction error: %w", res.DeliverTx.Error)
	}

	return nil
}

func (n *Node) getBlockStoreState(ctx context.Context) ([]std.Tx, error) {
	// get current genesis state
	genesis := n.GenesisDoc().AppState.(gnoland.GnoGenesisState)

	state := genesis.Txs[n.loadedPackages:] // ignore previously loaded packages
	lastBlock := n.getLatestBlockNumber()
	var blocnum uint64 = 1
	for ; blocnum <= lastBlock; blocnum++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		txs, txErr := n.GetBlockTransactions(blocnum)
		if txErr != nil {
			return nil, fmt.Errorf("unable to fetch block transactions, %w", txErr)
		}

		state = append(state, txs...)
	}

	// override current state
	return state, nil
}

func (n *Node) stopIfRunning() error {
	if n.Node != nil && n.Node.IsRunning() {
		if err := n.Node.Stop(); err != nil {
			return fmt.Errorf("unable to stop the node: %w", err)
		}
	}

	return nil
}

func (n *Node) reset(ctx context.Context, genesis gnoland.GnoGenesisState) error {
	// Setup node config
	nodeConfig := newNodeConfig(n.config.TMConfig, n.config.ChainID, genesis)
	nodeConfig.SkipFailingGenesisTxs = n.config.SkipFailingGenesisTxs
	nodeConfig.Genesis.ConsensusParams.Block.MaxGas = n.config.MaxGasPerBlock

	var recoverErr error

	// recoverFromError handles panics and converts them to errors.
	recoverFromError := func() {
		if r := recover(); r != nil {
			var ok bool
			if recoverErr, ok = r.(error); !ok {
				panic(r) // Re-panic if not an error.
			}
		}
	}

	// Execute node creation and handle any errors.
	defer recoverFromError()
	node, nodeErr := buildNode(n.logger, n.emitter, nodeConfig)
	if recoverErr != nil { // First check for recover error in case of panic
		return fmt.Errorf("recovered from a node panic: %w", recoverErr)
	}
	if nodeErr != nil { // Then for any node error
		return fmt.Errorf("unable to build the node: %w", nodeErr)
	}

	// Wait for the node to be ready
	select {
	case <-gnoland.GetNodeReadiness(node): // Ok
		n.Node = node
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func buildNode(logger *slog.Logger, emitter emitter.Emitter, cfg *gnoland.InMemoryNodeConfig) (*node.Node, error) {
	node, err := gnoland.NewInMemoryNode(logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create a new node: %w", err)
	}

	node.EventSwitch().AddListener("dev-emitter", func(evt tm2events.Event) {
		switch data := evt.(type) {
		case bft.EventTx:
			resEvt := events.TxResult{
				Height:   data.Result.Height,
				Index:    data.Result.Index,
				Response: data.Result.Response,
			}

			if err := amino.Unmarshal(data.Result.Tx, &resEvt.Tx); err != nil {
				logger.Error("unable to unwarp tx result",
					"error", err)
			}

			emitter.Emit(resEvt)
		}
	})

	if startErr := node.Start(); startErr != nil {
		return nil, fmt.Errorf("unable to start the node: %w", startErr)
	}

	return node, nil
}

func newNodeConfig(tmc *tmcfg.Config, chainid string, appstate gnoland.GnoGenesisState) *gnoland.InMemoryNodeConfig {
	// Create Mocked Identity
	pv := gnoland.NewMockedPrivValidator()
	genesis := gnoland.NewDefaultGenesisConfig(pv.GetPubKey(), chainid)
	genesis.AppState = appstate

	// Add self as validator
	self := pv.GetPubKey()
	genesis.Validators = []bft.GenesisValidator{
		{
			Address: self.Address(),
			PubKey:  self,
			Power:   10,
			Name:    "self",
		},
	}

	return &gnoland.InMemoryNodeConfig{
		PrivValidator:      pv,
		TMConfig:           tmc,
		Genesis:            genesis,
		GenesisMaxVMCycles: 10_000_000,
	}
}
