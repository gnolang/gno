package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type ShowSingleCfg struct {
	commonSingleCfg
}

// newShowSingleCmd creates the secrets show single command
func newShowSingleCmd(io commands.IO) *commands.Command {
	cfg := &ShowSingleCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "single",
			ShortUsage: "show single [flags]",
			ShortHelp:  "Shows required Gno secrets individually",
			LongHelp: "Shows the validator private key, the node p2p key and the validator's last sign state" +
				" at custom paths",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execShowSingle(cfg, io)
		},
	)

	return cmd
}

func (c *ShowSingleCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonSingleCfg.RegisterFlags(fs)
}

func execShowSingle(cfg *ShowSingleCfg, io commands.IO) error {
	var (
		validatorKeyPathSet   = cfg.validatorKeyPath != ""
		validatorStatePathSet = cfg.validatorStatePath != ""
		nodeKeyPathSet        = cfg.nodeKeyPath != ""
	)

	if !validatorKeyPathSet && !validatorStatePathSet && !nodeKeyPathSet {
		return errNoOutputSet
	}

	// Show the validator private key info, if any
	if validatorKeyPathSet {
		if err := readAndShowValidatorKey(cfg.validatorKeyPath, io); err != nil {
			return err
		}
	}

	// Show the last validator sign state, if any
	if validatorStatePathSet {
		if err := readAndShowValidatorState(cfg.validatorStatePath, io); err != nil {
			return err
		}
	}

	// Show the node key, if any
	if nodeKeyPathSet {
		if err := readAndShowNodeKey(cfg.nodeKeyPath, io); err != nil {
			return err
		}
	}

	return nil
}
