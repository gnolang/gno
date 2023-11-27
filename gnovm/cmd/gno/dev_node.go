package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	txclient "github.com/gnolang/tx-archive/backup/client"
)

const gnoDevChainID = "tendermint_test" // XXX: this is hardcoded and cannot be change bellow

var _ txclient.Client = (*DevNode)(nil)

type DevNode struct {
	muNode sync.Mutex
	node   *node.Node

	rootdir string
	logger  log.Logger

	pkgs  PkgsMap // path -> pkg
	state []std.Tx
}

var (
	defaultFee     = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
	defaultCreator = crypto.MustAddressFromString(integration.DefaultAccount_Address)
	defaultBalance = []gnoland.Balance{
		{
			Address: defaultCreator,
			Amount:  std.MustParseCoins("10000000000000ugnot"),
		},
	}
)

func NewDevNode(logger log.Logger, rootdir string, pkgslist []string) (*DevNode, error) {
	mpkgs, err := newPkgsMap(pkgslist)
	if err != nil {
		return nil, fmt.Errorf("unable map pkgs list: %w", err)
	}

	txs, err := mpkgs.Load(defaultCreator, defaultFee, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to load genesis packages: %w", err)
	}

	// generate genesis state
	genesis := gnoland.GnoGenesisState{
		Balances: defaultBalance,
		Txs:      txs,
	}

	node, err := newNode(logger, rootdir, genesis)
	if err != nil {
		return nil, fmt.Errorf("unable to create the node: %w", err)
	}

	if err := node.Start(); err != nil {
		return nil, fmt.Errorf("unable to start node: %w", err)
	}

	<-gnoland.WaitForNodeReadiness(node)

	return &DevNode{
		node:    node,
		pkgs:    mpkgs,
		rootdir: rootdir,
		logger:  logger,
	}, nil
}

func (d *DevNode) getLatestBlockNumber() uint64 {
	return uint64(d.node.BlockStore().Height())
}

func (d *DevNode) Close() error {
	d.muNode.Lock()
	defer d.muNode.Unlock()

	return d.node.Stop()
}

func (d *DevNode) ListPkgs() []gnomod.Pkg {
	return d.pkgs.toList()
}

func (d *DevNode) WaitForNodeReadiness() <-chan struct{} {
	return gnoland.WaitForNodeReadiness(d.node)
}

func (d *DevNode) GetRemoteAddress() string {
	return d.node.Config().RPC.ListenAddress
}

func (d *DevNode) UpdatePackages(paths ...string) error {
	for _, path := range paths {
		// list all packages from target path
		pkgslist, err := gnomod.ListPkgs(path)
		if err != nil {
			return fmt.Errorf("failed to list gno packages for %q: %w", path, err)
		}

		for _, pkg := range pkgslist {
			d.pkgs[pkg.Dir] = pkg
		}
	}

	return nil
}

func (d *DevNode) Reset() error {
	if d.node.IsRunning() {
		if err := d.node.Stop(); err != nil {
			return fmt.Errorf("unable to stop the node: %w", err)
		}
	}

	// generate genesis
	txs, err := d.pkgs.Load(defaultCreator, defaultFee, nil)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	genesis := gnoland.GnoGenesisState{
		Balances: defaultBalance,
		Txs:      txs,
	}

	return d.reset(genesis)
}

func (d *DevNode) ReloadAll(ctx context.Context) error {
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

func (d *DevNode) Reload(ctx context.Context) error {
	// save current (good) state
	state, err := d.saveState(ctx)
	if err != nil {
		return fmt.Errorf("unable to save state: %s", err.Error())
	}

	if d.node.IsRunning() {
		if err := d.node.Stop(); err != nil {
			return fmt.Errorf("unable to stop the node: %w", err)
		}
	}

	// generate genesis
	txs, err := d.pkgs.Load(defaultCreator, defaultFee, nil)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	genesis := gnoland.GnoGenesisState{
		Balances: defaultBalance,
		Txs:      txs,
	}

	// try to reset the node
	if err := d.reset(genesis); err != nil {
		return fmt.Errorf("unable to reset the node: %w", err)
	}

	resCh := make(chan abci.Response, 1)
	for _, tx := range state {
		if len(tx.Msgs) == 0 {
			continue
		}

		aminoTx, err := amino.Marshal(tx)
		if err != nil {
			return fmt.Errorf("unable to marshal transaction to amino binary, %w", err)
		}

		err = d.node.Mempool().CheckTx(aminoTx, func(res abci.Response) {
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
	}

	// ultimately restet state
	return nil
}

func (d *DevNode) reset(genesis gnoland.GnoGenesisState) error {
	var err error
	recoverError := func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				panic(r)
			}
		}
	}

	createNode := func() {
		defer recoverError()

		node, err := newNode(d.logger, d.rootdir, genesis)
		if err != nil {
			err = fmt.Errorf("unable to create node: %w", err)
			return
		}

		if err := node.Start(); err != nil {
			err = fmt.Errorf("unable to start the node: %w", err)
			return
		}

		d.node = node
	}

	createNode()

	<-d.WaitForNodeReadiness()

	return err
}

// GetBlockTransactions returns the transactions contained
// within the specified block, if any
func (d *DevNode) GetBlockTransactions(blockNum uint64) ([]std.Tx, error) {
	b := d.node.BlockStore().LoadBlock(int64(blockNum))
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
func (d *DevNode) GetLatestBlockNumber() (uint64, error) {
	return d.getLatestBlockNumber(), nil
}

func (n *DevNode) saveState(ctx context.Context) ([]std.Tx, error) {
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

// loadDefaultPackages loads the default packages for testing using a given creator address and gnoroot directory.
func loadPackagesFromDir(creator bft.Address, dir string) ([]std.Tx, error) {
	defaultFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
	txs, err := gnoland.LoadPackagesFromDir(dir, creator, defaultFee, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to load packages from %q: %w", dir, err)
	}

	return txs, nil
}

func newNode(logger log.Logger, rootdir string, genesis gnoland.GnoGenesisState) (*node.Node, error) {
	nodeConfig := gnoland.NewDefaultInMemoryNodeConfig(rootdir)
	nodeConfig.Genesis.AppState = genesis
	return gnoland.NewInMemoryNode(logger, nodeConfig)
}
