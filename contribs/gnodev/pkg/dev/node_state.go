package dev

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func (n *Node) SaveCurrentState(ctx context.Context) error {
	// Get current blockstore state
	state, err := n.getBlockStoreState(ctx)
	if err != nil {
		return fmt.Errorf("unable to save state: %s", err.Error())
	}

	n.initialState = state
	return nil
}

func (n *Node) ExportCurrentState(ctx context.Context) ([]std.Tx, error) {
	// Get current blockstore state
	state, err := n.getBlockStoreState(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to save state: %s", err.Error())
	}

	return state, nil
}

func (n *Node) MoveToPreviousTX(ctx context.Context) error {
	state := n.previousState
	if state == nil {
		var err error
		// Get current blockstore state
		state, err = n.getBlockStoreState(ctx)
		if err != nil {
			return fmt.Errorf("unable to save state: %s", err.Error())
		}
		n.stateIndex = len(state)
	}

	newIndex := n.stateIndex - 1
	if newIndex < 0 {
		return fmt.Errorf("not more previous state")
	}

	// Stop the node if it's currently running.
	if err := n.stopIfRunning(); err != nil {
		return fmt.Errorf("unable to stop the node: %w", err)
	}

	// Load genesis packages
	pkgsTxs, err := n.pkgs.Load(DefaultFee)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	// Create genesis with loaded pkgs + previous state
	newState := state[:newIndex]
	genesis := gnoland.GnoGenesisState{
		Balances: n.config.BalancesList,
		Txs:      append(pkgsTxs, newState...),
	}

	// Reset the node with the new genesis state.
	err = n.reset(ctx, genesis)
	n.logger.Info("reload done", "pkgs", len(pkgsTxs), "state applied", len(state))

	n.logger.Info("moving backward",
		"pkgs", len(pkgsTxs),
		"tx-index", newIndex,
		"state-applied", len(newState))

	// Update node infos
	n.stateIndex = newIndex
	if n.previousState == nil {
		n.previousState = state
	}

	n.loadedPackages = len(pkgsTxs)
	n.emitter.Emit(&events.Reload{})

	return nil
}

func (n *Node) MoveToNextTX(ctx context.Context) error {
	state := n.previousState
	if state == nil {
		return fmt.Errorf("already at the top of txs")
	}

	newIndex := n.stateIndex + 1
	if newIndex > len(state) {
		return fmt.Errorf("already at the top of txs")
	}

	// Stop the node if it's currently running.
	if err := n.stopIfRunning(); err != nil {
		return fmt.Errorf("unable to stop the node: %w", err)
	}

	// Load genesis packages
	pkgsTxs, err := n.pkgs.Load(DefaultFee)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	// Create genesis with loaded pkgs + previous state
	newState := state[:newIndex]
	genesis := gnoland.GnoGenesisState{
		Balances: n.config.BalancesList,
		Txs:      append(pkgsTxs, newState...),
	}

	// Reset the node with the new genesis state.
	err = n.reset(ctx, genesis)

	n.logger.Info("moving forward",
		"pkgs", len(pkgsTxs),
		"tx index", newIndex,
		"state applied", len(newState))

	// Update node infos
	n.stateIndex = newIndex
	n.loadedPackages = len(pkgsTxs)
	n.emitter.Emit(&events.Reload{})

	return nil
}

func (n *Node) ExportStateAsGenesis(ctx context.Context) (*bft.GenesisDoc, error) {
	// Get current blockstore state
	state, err := n.getBlockStoreState(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to save state: %s", err.Error())
	}

	// Get current blockstore state
	doc := *n.Node.GenesisDoc() // copy doc
	doc.AppState = gnoland.GnoGenesisState{
		Balances: n.config.BalancesList,
		Txs:      state,
	}

	return &doc, nil
}
