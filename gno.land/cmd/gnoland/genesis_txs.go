package main

import (
	"errors"
	"flag"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type txsCfg struct {
	commonCfg
}

var errInvalidGenesisStateType = errors.New("invalid genesis state type")

// newTxsCmd creates the genesis txs subcommand
func newTxsCmd(io commands.IO) *commands.Command {
	cfg := &txsCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "txs",
			ShortUsage: "txs <subcommand> [flags]",
			ShortHelp:  "manages the initial genesis transactions",
			LongHelp:   "Manages genesis transactions through input files",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newTxsAddCmd(cfg, io),
		newTxsRemoveCmd(cfg, io),
		newTxsExportCmd(cfg, io),
	)

	return cmd
}

func (c *txsCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonCfg.RegisterFlags(fs)
}

// appendGenesisTxs saves the given transactions to the genesis doc
func appendGenesisTxs(genesis *types.GenesisDoc, txs []std.Tx) error {
	// Initialize the app state if it's not present
	if genesis.AppState == nil {
		genesis.AppState = gnoland.GnoGenesisState{}
	}

	// Make sure the app state is the Gno genesis state
	state, ok := genesis.AppState.(gnoland.GnoGenesisState)
	if !ok {
		return errInvalidGenesisStateType
	}

	// Left merge the transactions
	fileTxStore := txStore(txs)
	genesisTxStore := txStore(state.Txs)

	// The genesis transactions have preference with the order
	// in the genesis.json
	if err := genesisTxStore.leftMerge(fileTxStore); err != nil {
		return err
	}

	// Save the state
	state.Txs = genesisTxStore
	genesis.AppState = state

	return nil
}
