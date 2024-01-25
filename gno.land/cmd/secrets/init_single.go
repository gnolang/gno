package main

import (
	"context"
	"errors"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errNoOutputSet = errors.New("no individual output path set")

type initSingleCfg struct {
	commonSingleCfg
}

// newInitSingleCmd creates the secrets init single command
func newInitSingleCmd(io commands.IO) *commands.Command {
	cfg := &initSingleCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "single",
			ShortUsage: "init single [flags]",
			ShortHelp:  "Initializes required Gno secrets individually",
			LongHelp: "Initializes the validator private key, the node p2p key and the validator's last sign state" +
				" at custom paths",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execInitSingle(cfg, io)
		},
	)

	return cmd
}

func (c *initSingleCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonSingleCfg.RegisterFlags(fs)
}

func execInitSingle(cfg *initSingleCfg, io commands.IO) error {
	var (
		validatorKeyPathSet   = cfg.validatorKeyPath != ""
		validatorStatePathSet = cfg.validatorStatePath != ""
		nodeKeyPathSet        = cfg.nodeKeyPath != ""
	)

	if !validatorKeyPathSet && !validatorStatePathSet && !nodeKeyPathSet {
		return errNoOutputSet
	}

	// Save the validator private key, if any
	if validatorKeyPathSet {
		if err := initAndSaveValidatorKey(cfg.validatorKeyPath, io); err != nil {
			return err
		}
	}

	// Save the last validator sign state, if any
	if validatorStatePathSet {
		if err := initAndSaveValidatorState(cfg.validatorStatePath, io); err != nil {
			return err
		}
	}

	// Save the node key, if any
	if nodeKeyPathSet {
		if err := initAndSaveNodeKey(cfg.nodeKeyPath, io); err != nil {
			return err
		}
	}

	return nil
}
