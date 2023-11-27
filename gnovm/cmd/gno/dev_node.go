package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	txclient "github.com/gnolang/tx-archive/backup/client"
)

const gnoDevChainID = "tendermint_test" // XXX: this is hardcoded and cannot be change bellow

var _ txclient.Client = (*DevNode)(nil)

type DevNode struct {
	muNode sync.Mutex

	rootdir string

	logger log.Logger
	node   *node.Node

	genesis gnoland.GnoGenesisState

	state []std.Tx
}

func NewDevNode(logger log.Logger, rootdir string, genesis gnoland.GnoGenesisState) (*DevNode, error) {
	node, err := newNode(logger, rootdir, genesis)
	if err != nil {
		return nil, fmt.Errorf("unable to create the node: %w", err)
	}

	if err := node.Start(); err != nil {
		return nil, fmt.Errorf("unable to start node: %w", err)
	}

	return &DevNode{
		node:    node,
		rootdir: rootdir,
		logger:  logger,
		genesis: genesis,
		state:   genesis.Txs,
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

func (d *DevNode) WaitForNodeReadiness() <-chan struct{} {
	return gnoland.WaitForNodeReadiness(d.node)
}

func (d *DevNode) GetRemoteAddress() string {
	return d.node.Config().RPC.ListenAddress
}

func (d *DevNode) Reset() error {
	if err := d.node.Stop(); err != nil {
		return fmt.Errorf("unable to stop the node: %w", err)
	}

	return d.reset(d.genesis)
}

func (d *DevNode) Reload(ctx context.Context) error {
	if err := d.saveState(ctx); err != nil {
		return fmt.Errorf("unable to save state: %s", err.Error())
	}

	d.logger.Debug("saved state", "txs", len(d.state))

	if err := d.node.Stop(); err != nil {
		return fmt.Errorf("unable to stop the node: %w", err)
	}

	genesis := gnoland.GnoGenesisState{
		Balances: d.genesis.Balances,
		Txs:      d.state,
	}

	return d.reset(genesis)
}

func (d *DevNode) reset(genesis gnoland.GnoGenesisState) error {
	d.logger.Debug("loading node", "state-txs", len(d.state))

	node, err := newNode(d.logger, d.rootdir, genesis)
	if err != nil {
		return fmt.Errorf("unable to create node: %w", err)
	}

	if err := node.Start(); err != nil {
		return fmt.Errorf("unable to start the node: %w", err)
	}

	d.node = node
	return nil
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

func (n *DevNode) saveState(ctx context.Context) error {
	lastBlock := n.getLatestBlockNumber()

	newState := make([]std.Tx, 0, int(lastBlock))
	var blocnum uint64 = 1
	for ; blocnum <= lastBlock; blocnum++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		txs, txErr := n.GetBlockTransactions(blocnum)
		if txErr != nil {
			return fmt.Errorf("unable to fetch block transactions, %w", txErr)
		}

		// Skip empty blocks
		if len(txs) == 0 {
			return nil
		}

		newState = append(newState, txs...)
	}

	// override current state
	n.state = newState
	return nil
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
