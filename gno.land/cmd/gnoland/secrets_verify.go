package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

type secretsVerifyCfg struct {
	commonAllCfg
}

// newSecretsVerifyCmd creates the secrets verify command
func newSecretsVerifyCmd(io commands.IO) *commands.Command {
	cfg := &secretsVerifyCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "secrets verify [flags] [<key>]",
			ShortHelp:  "verifies all Gno secrets in a common directory",
			LongHelp: fmt.Sprintf(
				"verifies the validator private key, the node p2p key and the validator's last sign state. "+
					"If a key is provided, it verifies the specified key value. Available keys: %s",
				getAvailableSecretsKeys(),
			),
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSecretsVerify(cfg, args, io)
		},
	)
}

func (c *secretsVerifyCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonAllCfg.RegisterFlags(fs)
}

func execSecretsVerify(cfg *secretsVerifyCfg, args []string, io commands.IO) error {
	// Make sure the directory is there
	if cfg.dataDir == "" || !isValidDirectory(cfg.dataDir) {
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

	// Construct the paths
	var (
		validatorKeyPath   = filepath.Join(cfg.dataDir, defaultValidatorKeyName)
		validatorStatePath = filepath.Join(cfg.dataDir, defaultValidatorStateName)
		nodeKeyPath        = filepath.Join(cfg.dataDir, defaultNodeKeyName)
	)

	switch key {
	case validatorPrivateKeyKey:
		// Validate the validator's private key
		_, err := signer.LoadFileKey(validatorKeyPath)

		return err
	case validatorStateKey:
		// Validate the validator's last sign state
		validatorState, err := state.LoadFileState(validatorStatePath)
		if err != nil {
			return err
		}

		// Attempt to read the validator key
		if validatorKey, err := signer.LoadFileKey(validatorKeyPath); validatorKey != nil && err == nil {
			// Validate the signature bytes
			return validateValidatorStateSignature(validatorState, validatorKey.PubKey)
		}
		io.Println("WARN: Skipped verification of validator state, as validator key is not present")

		return nil
	case nodeIDKey:
		return readAndVerifyNodeKey(nodeKeyPath, io)
	default:
		// Validate the validator's private key
		validatorKey, err := signer.LoadFileKey(validatorKeyPath)
		if err != nil {
			return err
		}

		// Validate the validator's last sign state
		validatorState, err := state.LoadFileState(validatorStatePath)
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
}

// validateValidatorStateSignature validates the signature of the given validator state.
func validateValidatorStateSignature(validatorState *state.FileState, pubKey crypto.PubKey) error {
	if !pubKey.VerifyBytes(validatorState.SignBytes, validatorState.Signature) {
		return fmt.Errorf("validator state signature is invalid")
	}

	return nil
}

// readAndVerifyNodeKey reads the node p2p key from the given path and verifies it
func readAndVerifyNodeKey(path string, io commands.IO) error {
	nodeKey, err := readNodeKey(path)
	if err != nil {
		return fmt.Errorf("unable to read node p2p key, %w", err)
	}

	if err := validateNodeKey(nodeKey); err != nil {
		return err
	}

	io.Printfln("Node P2P key at %s is valid", path)

	return nil
}
