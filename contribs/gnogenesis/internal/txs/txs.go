package txs

import (
	"errors"
	"flag"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type txsCfg struct {
	common.Cfg
}

var errInvalidGenesisStateType = errors.New("invalid genesis state type")

// NewTxsCmd creates the genesis txs subcommand
func NewTxsCmd(io commands.IO) *commands.Command {
	cfg := &txsCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "txs",
			ShortUsage: "<subcommand> [flags]",
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
		newTxsListCmd(cfg, io),
	)

	return cmd
}

func (c *txsCfg) RegisterFlags(fs *flag.FlagSet) {
	c.Cfg.RegisterFlags(fs)
}

// appendGenesisTxs saves the given transactions to the genesis doc
func appendGenesisTxs(genesis *types.GenesisDoc, txs []gnoland.TxWithMetadata) error {
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

// txStore is a wrapper for TM2 transactions
type txStore []gnoland.TxWithMetadata

// leftMerge merges the two tx stores, with
// preference to the left
func (i *txStore) leftMerge(b txStore) error {
	// Build out the tx hash map
	txHashMap := make(map[string]struct{}, len(*i))

	for _, tx := range *i {
		txHash, err := getTxHash(tx.Tx)
		if err != nil {
			return err
		}

		txHashMap[txHash] = struct{}{}
	}

	for _, tx := range b {
		txHash, err := getTxHash(tx.Tx)
		if err != nil {
			return err
		}

		if _, exists := txHashMap[txHash]; !exists {
			*i = append(*i, tx)
		}
	}

	return nil
}
