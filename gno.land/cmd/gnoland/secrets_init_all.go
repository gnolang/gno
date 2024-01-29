package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type initAllCfg struct {
	commonAllCfg
}

// newInitAllCmd creates the secrets init all command
func newInitAllCmd(io commands.IO) *commands.Command {
	cfg := &initAllCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "all",
			ShortUsage: "secrets init all [flags]",
			ShortHelp:  "Initializes required Gno secrets in a common directory",
			LongHelp:   "Initializes the validator private key, the node p2p key and the validator's last sign state",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execInitAll(cfg, io)
		},
	)

	return cmd
}

func (c *initAllCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonAllCfg.RegisterFlags(fs)
}

func execInitAll(cfg *initAllCfg, io commands.IO) error {
	// Check the data output directory path
	if cfg.dataDir == "" {
		return errInvalidDataDir
	}

	// Make sure the directory is there
	if err := os.MkdirAll(cfg.dataDir, 0o755); err != nil {
		return fmt.Errorf("unable to create secrets dir, %w", err)
	}

	// Construct the paths
	var (
		validatorKeyPath   = filepath.Join(cfg.dataDir, defaultValidatorKeyName)
		validatorStatePath = filepath.Join(cfg.dataDir, defaultValidatorStateName)
		nodeKeyPath        = filepath.Join(cfg.dataDir, defaultNodeKeyName)
	)

	// Initialize and save the validator's private key
	if err := initAndSaveValidatorKey(validatorKeyPath, io); err != nil {
		return err
	}

	// Initialize and save the validator's last sign state
	if err := initAndSaveValidatorState(validatorStatePath, io); err != nil {
		return err
	}

	// Initialize and save the node's p2p key
	return initAndSaveNodeKey(nodeKeyPath, io)
}
