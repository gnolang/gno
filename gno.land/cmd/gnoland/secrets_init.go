package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

var errOverwriteNotEnabled = errors.New("overwrite not enabled")

type secretsInitCfg struct {
	commonAllCfg

	forceOverwrite bool
	keyType        string
}

// newSecretsInitCmd creates the secrets init command
func newSecretsInitCmd(io commands.IO) *commands.Command {
	cfg := &secretsInitCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "secrets init [flags] [<key>]",
			ShortHelp:  "initializes required Gno secrets in a common directory",
			LongHelp: fmt.Sprintf(
				"initializes the validator private key, the node p2p key and the validator's last sign state. "+
					"If a key is provided, it initializes the specified key. Available keys: %s",
				getAvailableSecretsKeys(),
			),
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSecretsInit(cfg, args, io)
		},
	)
}

func (c *secretsInitCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonAllCfg.RegisterFlags(fs)

	fs.BoolVar(
		&c.forceOverwrite,
		"force",
		false,
		"overwrite existing secrets, if any",
	)

	fs.StringVar(
		&c.keyType,
		"key-type",
		string(signer.DefaultKeyType),
		fmt.Sprintf("validator key signing scheme (%q or %q)",
			signer.KeyTypeEd25519, signer.KeyTypeSecp256k1),
	)
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

	// Parse the validator key type up-front so the user gets an early
	// error on a typo before we touch the filesystem. An empty string
	// means "use the default scheme" — programmatic callers (e.g. the
	// lazy-init path in gnoland start) construct secretsInitCfg directly
	// without going through flag registration.
	keyType := signer.DefaultKeyType
	if cfg.keyType != "" {
		parsed, err := signer.ParseKeyType(cfg.keyType)
		if err != nil {
			return err
		}
		keyType = parsed
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

	initValidatorKey := func(path string, io commands.IO) error {
		return initAndSaveValidatorKeyOfType(path, keyType, io)
	}

	switch key {
	case validatorPrivateKeyKey:
		if osm.FileExists(validatorKeyPath) && !cfg.forceOverwrite {
			return fmt.Errorf("unable to overwrite validator key, %w", errOverwriteNotEnabled)
		}

		// Initialize and save the validator's private key
		return initValidatorKey(validatorKeyPath, io)
	case nodeIDKey:
		if osm.FileExists(nodeKeyPath) && !cfg.forceOverwrite {
			return fmt.Errorf("unable to overwrite the node' p2p key, %w", errOverwriteNotEnabled)
		}

		// Initialize and save the node's p2p key
		return initAndSaveNodeKey(nodeKeyPath, io)
	case validatorStateKey:
		if osm.FileExists(validatorStatePath) && !cfg.forceOverwrite {
			return fmt.Errorf("unable to overwrite validator last sign state, %w", errOverwriteNotEnabled)
		}

		// Initialize and save the validator's last sign state
		return initAndSaveValidatorState(validatorStatePath, io)
	default:
		// No key provided, initialize everything
		return errors.Join(
			overwriteGuard(validatorKeyPath, initValidatorKey, cfg.forceOverwrite, io),
			overwriteGuard(validatorStatePath, initAndSaveValidatorState, cfg.forceOverwrite, io),
			overwriteGuard(nodeKeyPath, initAndSaveNodeKey, cfg.forceOverwrite, io),
		)
	}
}

// overwriteGuard guards against unwanted secret overwrites,
// and executes the secret initialization if the secret is not present
func overwriteGuard(
	path string,
	initFn func(string, commands.IO) error,
	overwriteEnabled bool,
	io commands.IO,
) error {
	// Check if the secret already exists
	if osm.FileExists(path) && !overwriteEnabled {
		return fmt.Errorf(
			"unable to overwrite secret at %q, %w",
			path,
			errOverwriteNotEnabled,
		)
	}

	// Secret doesn't exist, initialize it
	return initFn(path, io)
}

// initAndSaveValidatorKeyOfType generates a validator private key of the
// given scheme and saves it to the given path.
func initAndSaveValidatorKeyOfType(path string, keyType signer.KeyType, io commands.IO) error {
	if _, err := signer.GeneratePersistedFileKeyOfType(path, keyType); err != nil {
		return fmt.Errorf("unable to save validator key, %w", err)
	}

	io.Printfln("Validator private key (%s) saved at %s", keyType, path)

	return nil
}

// initAndSaveValidatorState generates an empty last validator sign state and saves it to the given path
func initAndSaveValidatorState(path string, io commands.IO) error {
	// Initialize the validator's last sign state
	if _, err := fstate.GeneratePersistedFileState(path); err != nil {
		return fmt.Errorf("unable to save last validator sign state, %w", err)
	}

	io.Printfln("Validator last sign state saved at %s", path)

	return nil
}

// initAndSaveNodeKey generates a node p2p key and saves it to the given path
func initAndSaveNodeKey(path string, io commands.IO) error {
	// Initialize the node's p2p key
	if _, err := types.GeneratePersistedNodeKey(path); err != nil {
		return fmt.Errorf("unable to save node p2p key, %w", err)
	}

	io.Printfln("Node key saved at %s", path)

	return nil
}
