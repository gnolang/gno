// Dedicated to my love, Lexi.
package keyscli

import (
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
)

func NewRootCmd(io commands.IO, base client.BaseOptions) *commands.Command {
	cfg := &client.BaseCfg{
		BaseOptions: base,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "gno.land keychain & client",
			Options: []ff.Option{
				ff.WithConfigFileFlag("config"),
				ff.WithConfigFileParser(fftoml.Parser),
			},
		},
		cfg,
		commands.HelpExec,
	)

	// OnTxSuccess is only used by NewBroadcastCmd
	cfg.OnTxSuccess = func(tx std.Tx, res *ctypes.ResultBroadcastTxCommit) {
		client.PrintTxInfo(tx, res, io)
	}
	cmd.AddSubCommands(
		client.NewAddCmd(cfg, io),
		client.NewDeleteCmd(cfg, io),
		client.NewRotateCmd(cfg, io),
		client.NewGenerateCmd(cfg, io),
		client.NewExportCmd(cfg, io),
		client.NewImportCmd(cfg, io),
		client.NewListCmd(cfg, io),
		client.NewSignCmd(cfg, io),
		client.NewVerifyCmd(cfg, io),
		client.NewQueryCmd(cfg, io),
		client.NewBroadcastCmd(cfg, io),
		client.NewMultisignCmd(cfg, io),
		client.NewVersionCmd(cfg, io),

		// Custom MakeTX command
		NewMakeTxCmd(cfg, io),
	)

	return cmd
}
