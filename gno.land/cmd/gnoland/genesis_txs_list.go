package main

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
)

// newTxsListCmd list all transactions on the specified genesis file
func newTxsListCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "txs list [flags] [<arg>...]",
			ShortHelp:  "lists transactions existing on genesis.json",
			LongHelp:   "Lists transactions existing on genesis.json",
		},
		commands.NewEmptyConfig(),
		func(ctx context.Context, args []string) error {
			return execTxsListCmd(io, txsCfg)
		},
	)

	return cmd
}

func execTxsListCmd(io commands.IO, cfg *txsCfg) error {
	genesis, err := types.GenesisDocFromFile(cfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	gs, ok := genesis.AppState.(gnoland.GnoGenesisState)
	if !ok {
		return fmt.Errorf("genesis state is not using the correct Gno Genesis type")
	}

	tw := tabwriter.NewWriter(io.Out(), 0, 8, 2, '\t', 0)
	for _, tx := range gs.Txs {
		hash, err := getTxHash(tx)
		if err != nil {
			return fmt.Errorf("unable to generate tx hash, %w", err)
		}
		for _, msg := range tx.Msgs {
			switch m := msg.(type) {
			case vmm.MsgAddPackage:
				fmt.Fprintf(tw, "tx:%s\ttype:create\tpath:%s\tfiles:%d\tcreator:%s\t\n", hash, m.Package.Path, len(m.Package.Files), m.Creator.String())
			case vmm.MsgCall:
				fmt.Fprintf(tw, "tx:%s\ttype:call\tpath:%s\tparams:%d\tcaller:%s\t\n", hash, m.PkgPath, len(m.Args), m.Caller.String())
			case bank.MsgSend:
				fmt.Fprintf(tw, "tx:%s\ttype:send\tfrom:%s\tto:%s\tamount:%s\t\n", hash, m.FromAddress.String(), m.ToAddress.String(), m.Amount.String())
			}
		}
	}

	return tw.Flush()
}
