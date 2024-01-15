package dev

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	//backup "github.com/gnolang/tx-archive/backup/client"
	//restore "github.com/gnolang/tx-archive/restore/client"
)

const gnoDevChainID = "tendermint_test" // XXX: this is hardcoded and cannot be change bellow

// Node is not thread safe
type Node struct {
	*node.Node

	logger log.Logger
	pkgs   PkgsMap // path -> pkg
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

func NewDevNode(ctx context.Context, logger log.Logger, pkgslist []string) (*Node, error) {
	mpkgs, err := newPkgsMap(pkgslist)
	if err != nil {
		return nil, fmt.Errorf("unable map pkgs list: %w", err)
	}

	txs, err := mpkgs.Load(DefaultCreator, DefaultFee, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to load genesis packages: %w", err)
	}

	// generate genesis state
	genesis := gnoland.GnoGenesisState{
		Balances: DefaultBalance,
		Txs:      txs,
	}

	node, err := newNode(logger, genesis)
	if err != nil {
		return nil, fmt.Errorf("unable to create the node: %w", err)
	}

	if err := node.Start(); err != nil {
		return nil, fmt.Errorf("unable to start node: %w", err)
	}

	// Wait for readiness
	select {
	case <-gnoland.GetNodeReadiness(node): // ok
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return &Node{
		Node:   node,
		pkgs:   mpkgs,
		logger: logger,
	}, nil
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
	if d.Node.IsRunning() {
		if err := d.Node.Stop(); err != nil {
			return fmt.Errorf("unable to stop the node: %w", err)
		}
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
	return d.reset(ctx, genesis)
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
	// Save the current state of the node.
	state, err := d.saveState(ctx)
	if err != nil {
		return fmt.Errorf("unable to save state: %s", err.Error())
	}

	// Stop the node if it's currently running.
	if d.Node.IsRunning() {
		if err := d.Node.Stop(); err != nil {
			return fmt.Errorf("unable to stop the node: %w", err)
		}
	}

	// Generate a new genesis state based on the current packages.
	txs, err := d.pkgs.Load(DefaultCreator, DefaultFee, nil)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	genesis := gnoland.GnoGenesisState{
		Balances: DefaultBalance,
		Txs:      txs,
	}

	// Reset the node with the new genesis state.
	if err := d.reset(ctx, genesis); err != nil {
		return fmt.Errorf("unable to reset the node: %w", err)
	}

	// Attempt to resend transactions from the saved state.
	for _, tx := range state {
		if len(tx.Msgs) == 0 { // Skip empty transactions.
			continue
		}

		if err := d.SendTransaction(&tx); err != nil {
			return fmt.Errorf("unable to send transaction: %w", err)
		}
	}

	return nil
}

func (d *Node) reset(ctx context.Context, genesis gnoland.GnoGenesisState) error {
	var err error

	// recoverError handles panics and converts them to errors.
	recoverError := func() {
		if r := recover(); r != nil {
			panicErr, ok := r.(error)
			if !ok {
				panic(r) // Re-panic if not an error.
			}

			err = panicErr
		}
	}

	createNode := func() {
		defer recoverError()

		node, nodeErr := newNode(d.logger, genesis)
		if nodeErr != nil {
			err = fmt.Errorf("unable to create node: %w", nodeErr)
			return
		}

		if startErr := node.Start(); startErr != nil {
			err = fmt.Errorf("unable to start the node: %w", startErr)
			return
		}

		d.Node = node
	}

	// Execute node creation and handle any errors.
	createNode()
	if err != nil {
		return err
	}

	// Wait for the node to be ready
	select {
	case <-d.GetNodeReadiness(): // Ok
	case <-ctx.Done():
		return ctx.Err()
	}

	return err
}

// GetBlockTransactions returns the transactions contained
// within the specified block, if any
func (d *Node) GetBlockTransactions(blockNum uint64) ([]std.Tx, error) {
	b := d.Node.BlockStore().LoadBlock(int64(blockNum))
	txs := make([]std.Tx, len(b.Txs))
	for i, encodedTx := range b.Txs {
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

// SendTransaction executes a broadcast sync send
// of the specified transaction to the chain
func (d *Node) SendTransaction(tx *std.Tx) error {
	resCh := make(chan abci.Response, 1)

	aminoTx, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal transaction to amino binary, %w", err)
	}

	err = d.Node.Mempool().CheckTx(aminoTx, func(res abci.Response) {
		resCh <- res
	})
	if err != nil {
		return fmt.Errorf("unable to check tx: %w", err)
	}

	res := <-resCh
	r := res.(abci.ResponseCheckTx)
	if r.Error != nil {
		return fmt.Errorf("unable to broadcast tx: %w\nLog: %s", r.Error, r.Log)
	}

	return nil
}

func (n *Node) saveState(ctx context.Context) ([]std.Tx, error) {
	lastBlock := n.getLatestBlockNumber()

	state := make([]std.Tx, 0, int(lastBlock))
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

		// Skip empty blocks
		state = append(state, txs...)
	}

	// override current state
	return state, nil
}

type PkgsMap map[string]gnomod.Pkg

func newPkgsMap(paths []string) (PkgsMap, error) {
	pkgs := make(map[string]gnomod.Pkg)
	for _, path := range paths {
		// list all packages from target path
		pkgslist, err := gnomod.ListPkgs(path)
		if err != nil {
			return nil, fmt.Errorf("listing gno packages: %w", err)
		}

		for _, pkg := range pkgslist {
			if pkg.Dir == "" {
				continue
			}

			if _, ok := pkgs[pkg.Dir]; ok {
				continue // skip
			}
			pkgs[pkg.Dir] = pkg
		}
	}

	// Filter out draft packages.
	return pkgs, nil
}

func (pm PkgsMap) toList() gnomod.PkgList {
	list := make([]gnomod.Pkg, 0, len(pm))
	for _, pkg := range pm {
		list = append(list, pkg)
	}
	return list
}

func (pm PkgsMap) Load(creator bft.Address, fee std.Fee, deposit std.Coins) ([]std.Tx, error) {
	pkgs := pm.toList()

	sorted, err := pkgs.Sort()
	if err != nil {
		return nil, fmt.Errorf("unable to sort pkgs: %w", err)
	}

	nonDraft := sorted.GetNonDraftPkgs()
	txs := []std.Tx{}
	for _, pkg := range nonDraft {
		// Open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)
		if err := memPkg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid package: %w", err)
		}

		// Create transaction
		tx := std.Tx{
			Fee: fee,
			Msgs: []std.Msg{
				vmm.MsgAddPackage{
					Creator: creator,
					Package: memPkg,
					Deposit: deposit,
				},
			},
		}

		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}

	return txs, nil
}

func newNode(logger log.Logger, genesis gnoland.GnoGenesisState) (*node.Node, error) {
	rootdir := gnoenv.RootDir()

	nodeConfig := gnoland.NewDefaultInMemoryNodeConfig(rootdir)
	nodeConfig.SkipFailingGenesisTxs = true
	nodeConfig.TMConfig.Consensus.SkipTimeoutCommit = false // avoid time drifting, see issue #1507

	nodeConfig.Genesis.AppState = genesis
	return gnoland.NewInMemoryNode(logger, nodeConfig)
}
