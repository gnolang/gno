package balances

import (
	"flag"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type balancesCfg struct {
	common.Cfg
}

// NewBalancesCmd creates the genesis balances subcommand
func NewBalancesCmd(io commands.IO) *commands.Command {
	cfg := &balancesCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "balances",
			ShortUsage: "<subcommand> [flags]",
			ShortHelp:  "manages genesis.json account balances",
			LongHelp:   "Manipulates the initial genesis.json account balances (pre-mines)",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newBalancesAddCmd(cfg, io),
		newBalancesRemoveCmd(cfg, io),
		newBalancesExportCmd(cfg, io),
	)

	return cmd
}

func (c *balancesCfg) RegisterFlags(fs *flag.FlagSet) {
	c.Cfg.RegisterFlags(fs)
}
