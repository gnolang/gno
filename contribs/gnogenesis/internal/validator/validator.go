package validator

import (
	"flag"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type validatorCfg struct {
	common.Cfg

	address string
}

// NewValidatorCmd creates the genesis validator subcommand
func NewValidatorCmd(io commands.IO) *commands.Command {
	cfg := &validatorCfg{
		Cfg: common.Cfg{},
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "validator",
			ShortUsage: "<subcommand> [flags]",
			ShortHelp:  "validator set management in genesis.json",
			LongHelp:   "Manipulates the genesis.json validator set",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newValidatorAddCmd(cfg, io),
		newValidatorRemoveCmd(cfg, io),
	)

	return cmd
}

func (c *validatorCfg) RegisterFlags(fs *flag.FlagSet) {
	c.Cfg.RegisterFlags(fs)

	fs.StringVar(
		&c.address,
		"address",
		"",
		"the gno bech32 address of the validator",
	)
}
