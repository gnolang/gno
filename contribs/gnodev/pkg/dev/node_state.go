package dev

import (
	"context"
	"errors"
	"fmt"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var ErrEmptyState = errors.New("empty state")

// Save the current state as initialState
func (n *Node) SaveCurrentState(ctx context.Context) error {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	// Get current blockstore state
	state, err := n.getState(ctx)
	if err != nil {
		return fmt.Errorf("unable to save state: %s", err.Error())
	}

	n.initialState = state[:n.currentStateIndex]
	return nil
}

// Export the current state as list of txs
func (n *Node) ExportCurrentState(ctx context.Context) ([]std.Tx, error) {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	// Get current blockstore state
	state, err := n.getState(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to save state: %s", err.Error())
	}

	return state[:n.currentStateIndex], nil
}

func (n *Node) getState(ctx context.Context) ([]std.Tx, error) {
	if n.state == nil {
		var err error
		n.state, err = n.getBlockStoreState(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to save state: %s", err.Error())
		}
	}

	return n.state, nil
}

// MoveBy adjusts the current state of the node by `x` transactions.
// `x` can be negative to move backward or positive to move forward, however, index boundaries are respected
// with a lower limit of 0 and upper limit equaling the total number of states.
// If a move is successful, node is reloaded.
func (n *Node) MoveBy(ctx context.Context, x int) error {
	n.muNode.Lock()
	defer n.muNode.Unlock()

	newIndex := n.currentStateIndex + x
	state, err := n.getState(ctx)
	if err != nil {
		return fmt.Errorf("unable to get current state: %w", err)
	}

	maxState := len(state)
	switch {
	case maxState == 0: // no state
		return ErrEmptyState
	case newIndex < 0:
		newIndex = 0
		n.logger.Info("minimum state reached", "tx-index", fmt.Sprintf("%d/%d", newIndex, maxState))
	case newIndex > maxState:
		newIndex = maxState
		n.logger.Info("maximum state reached", "tx-index", fmt.Sprintf("%d/%d", newIndex, maxState))
	}

	if newIndex == n.currentStateIndex {
		return nil
	}

	// Load genesis packages
	pkgsTxs, err := n.pkgs.Load(DefaultFee)
	if err != nil {
		return fmt.Errorf("unable to load pkgs: %w", err)
	}

	newState := n.state[:newIndex]

	// Create genesis with loaded pkgs + previous state
	genesis := gnoland.GnoGenesisState{
		Balances: n.config.BalancesList,
		Txs:      append(pkgsTxs, newState...),
	}

	// Reset the node with the new genesis state.
	err = n.rebuildNode(ctx, genesis)
	if err != nil {
		return fmt.Errorf("uanble to rebuild node: %w", err)
	}

	n.logger.Info("moving to", "tx-index", fmt.Sprintf("%d/%d", newIndex, maxState))

	// Update node infos
	n.currentStateIndex = newIndex
	n.emitter.Emit(&events.Reload{})

	return nil
}

func (n *Node) MoveToPreviousTX(ctx context.Context) error {
	return n.MoveBy(ctx, -1)
}

func (n *Node) MoveToNextTX(ctx context.Context) error {
	return n.MoveBy(ctx, 1)
}

// Export the current state as genesis doc
func (n *Node) ExportStateAsGenesis(ctx context.Context) (*bft.GenesisDoc, error) {
	n.muNode.RLock()
	defer n.muNode.RUnlock()

	// Get current blockstore state
	state, err := n.getState(ctx)
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
