package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type secretsInitCfg struct {
	commonAllCfg
}

// newSecretsInitCmd creates the secrets init command
func newSecretsInitCmd(io commands.IO) *commands.Command {
	cfg := &secretsInitCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "secrets init [flags] [<key>]",
			ShortHelp:  "initializes required Gno secrets in a common directory",
			LongHelp:   "initializes the validator private key, the node p2p key and the validator's last sign state",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSecretsInit(cfg, args, io)
		},
	)
}

func (c *secretsInitCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonAllCfg.RegisterFlags(fs)
}

func execSecretsInit(cfg *secretsInitCfg, args []string, io commands.IO) error {
	// Check the data output directory path
	if cfg.dataDir == "" {
		return errInvalidDataDir
	}

	// Verify the secrets key
	if err := verifySecretsKey(args); err != nil {
		return err
	}

	var key string

	if len(args) > 0 {
		key = args[0]
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

	switch key {
	case validatorPrivateKeyKey:
		// Initialize and save the validator's private key
		return initAndSaveValidatorKey(validatorKeyPath, io)
	case nodeKeyKey:
		// Initialize and save the node's p2p key
		return initAndSaveNodeKey(nodeKeyPath, io)
	case validatorStateKey:
		// Initialize and save the validator's last sign state
		return initAndSaveValidatorState(validatorStatePath, io)
	default:
		// No key provided, initialize everything
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
}

// initAndSaveValidatorKey generates a validator private key and saves it to the given path
func initAndSaveValidatorKey(path string, io commands.IO) error {
	// Initialize the validator's private key
	privateKey := generateValidatorPrivateKey()

	// Save the key
	if err := saveSecretData(privateKey, path); err != nil {
		return fmt.Errorf("unable to save validator key, %w", err)
	}

	io.Printfln("Validator private key saved at %s", path)

	return nil
}

// initAndSaveValidatorState generates an empty last validator sign state and saves it to the given path
func initAndSaveValidatorState(path string, io commands.IO) error {
	// Initialize the validator's last sign state
	validatorState := generateLastSignValidatorState()

	// Save the last sign state
	if err := saveSecretData(validatorState, path); err != nil {
		return fmt.Errorf("unable to save last validator sign state, %w", err)
	}

	io.Printfln("Validator last sign state saved at %s", path)

	return nil
}

// initAndSaveNodeKey generates a node p2p key and saves it to the given path
func initAndSaveNodeKey(path string, io commands.IO) error {
	// Initialize the node's p2p key
	nodeKey := generateNodeKey()

	// Save the node key
	if err := saveSecretData(nodeKey, path); err != nil {
		return fmt.Errorf("unable to save node p2p key, %w", err)
	}

	io.Printfln("Node key saved at %s", path)

	return nil
}
