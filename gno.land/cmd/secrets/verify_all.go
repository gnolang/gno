package main

import (
	"context"
	"flag"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type verifyAllCfg struct {
	commonAllCfg
}

// newVerifyAllCmd creates the secrets verify all command
func newVerifyAllCmd(io commands.IO) *commands.Command {
	cfg := &verifyAllCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "all",
			ShortUsage: "verify all [flags]",
			ShortHelp:  "Verifies all Gno secrets in a common directory",
			LongHelp:   "Verifies the validator private key, the node p2p key and the validator's last sign state",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execVerifyAll(cfg, io)
		},
	)

	return cmd
}

func (c *verifyAllCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonAllCfg.RegisterFlags(fs)
}

func execVerifyAll(cfg *verifyAllCfg, io commands.IO) error {
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

	// Validate the validator's private key
	validatorKey, err := readAndVerifyValidatorKey(validatorKeyPath, io)
	if err != nil {
		return err
	}

	// Validate the validator's last sign state
	validatorState, err := readAndVerifyValidatorState(validatorStatePath, io)
	if err != nil {
		return err
	}

	// Validate the signature bytes
	if err = validateValidatorStateSignature(validatorState, validatorKey.PubKey); err != nil {
		return err
	}

	// Validate the node's p2p key
	return readAndVerifyNodeKey(nodeKeyPath, io)
}
