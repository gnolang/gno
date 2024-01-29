package main

import (
	"context"
	"flag"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type showAllCfg struct {
	commonAllCfg
}

// newShowAllCmd creates the secrets show all command
func newShowAllCmd(io commands.IO) *commands.Command {
	cfg := &showAllCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "all",
			ShortUsage: "secrets show all [flags]",
			ShortHelp:  "Shows all Gno secrets in a common directory",
			LongHelp:   "Shows the validator private key, the node p2p key and the validator's last sign state",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execShowAll(cfg, io)
		},
	)

	return cmd
}

func (c *showAllCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonAllCfg.RegisterFlags(fs)
}

func execShowAll(cfg *showAllCfg, io commands.IO) error {
	// Make sure the directory is there
	if cfg.dataDir == "" || !isValidDirectory(cfg.dataDir) {
		return errInvalidDataDir
	}

	// Construct the paths
	var (
		validatorKeyPath   = filepath.Join(cfg.dataDir, defaultValidatorKeyName)
		validatorStatePath = filepath.Join(cfg.dataDir, defaultValidatorStateName)
		nodeKeyPath        = filepath.Join(cfg.dataDir, defaultNodeKeyName)
	)

	// Show the node's p2p info
	if err := readAndShowNodeKey(nodeKeyPath, io); err != nil {
		return err
	}

	// Show the validator's key info
	if err := readAndShowValidatorKey(validatorKeyPath, io); err != nil {
		return err
	}

	// Show the validator's last sign state
	return readAndShowValidatorState(validatorStatePath, io)
}
