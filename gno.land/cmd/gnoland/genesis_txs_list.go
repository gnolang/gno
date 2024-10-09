package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var ErrWrongGenesisType = errors.New("genesis state is not using the correct Gno Genesis type")

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
	genesis, err := types.GenesisDocFromFile(cfg.homeDir.GenesisFilePath())
	if err != nil {
		return fmt.Errorf("%w, %w", errUnableToLoadGenesis, err)
	}

	gs, ok := genesis.AppState.(gnoland.GnoGenesisState)
	if !ok {
		return ErrWrongGenesisType
	}

	b, err := amino.MarshalJSONIndent(gs.Txs, "", "    ")
	if err != nil {
		return errors.New("error marshalling data to amino JSON")
	}

	buf := bytes.NewBuffer(b)
	_, err = buf.WriteTo(io.Out())

	return err
}
