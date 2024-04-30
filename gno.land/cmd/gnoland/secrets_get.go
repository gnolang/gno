package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

type secretsGetCfg struct {
	commonAllCfg
}

// newSecretsGetCmd creates the secrets get command
func newSecretsGetCmd(io commands.IO) *commands.Command {
	cfg := &secretsGetCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "get",
			ShortUsage: "secrets get [flags] [<key>]",
			ShortHelp:  "shows all Gno secrets present in a common directory",
			LongHelp: fmt.Sprintf(
				"shows the validator private key, the node p2p key and the validator's last sign state. "+
					"If a key is provided, it shows the specified key value. Available keys: %s",
				getAvailableSecretsKeys(),
			),
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSecretsGet(cfg, args, io)
		},
	)

	return cmd
}

func (c *secretsGetCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonAllCfg.RegisterFlags(fs)
}

func execSecretsGet(cfg *secretsGetCfg, args []string, io commands.IO) error {
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
		// Show the validator's key info
		return readAndShowValidatorKey(validatorKeyPath, io)
	case validatorStateKey:
		// Show the validator's last sign state
		return readAndShowValidatorState(validatorStatePath, io)
	case nodeKeyKey:
		// Show the node's p2p info
		return readAndShowNodeKey(nodeKeyPath, io)
	default:
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
}

// readAndShowValidatorKey reads and shows the validator key from the given path
func readAndShowValidatorKey(path string, io commands.IO) error {
	validatorKey, err := readSecretData[privval.FilePVKey](path)
	if err != nil {
		return fmt.Errorf("unable to read validator key, %w", err)
	}

	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Validator Key Info]\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Address:\t%s\n", validatorKey.Address.String()); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Public Key:\t%s\n", validatorKey.PubKey.String()); err != nil {
		return err
	}

	return w.Flush()
}

// readAndShowValidatorState reads and shows the validator state from the given path
func readAndShowValidatorState(path string, io commands.IO) error {
	validatorState, err := readSecretData[privval.FilePVLastSignState](path)
	if err != nil {
		return fmt.Errorf("unable to read validator state, %w", err)
	}

	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Last Validator Sign State Info]\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Height:\t%d\n",
		validatorState.Height,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Round:\t%d\n",
		validatorState.Round,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Step:\t%d\n",
		validatorState.Step,
	); err != nil {
		return err
	}

	if validatorState.Signature != nil {
		if _, err := fmt.Fprintf(
			w,
			"Signature:\t%X\n",
			validatorState.Signature,
		); err != nil {
			return err
		}
	}

	if validatorState.SignBytes != nil {
		if _, err := fmt.Fprintf(
			w,
			"Sign Bytes:\t%X\n",
			validatorState.SignBytes,
		); err != nil {
			return err
		}
	}

	return w.Flush()
}

// readAndShowNodeKey reads and shows the node p2p key from the given path
func readAndShowNodeKey(path string, io commands.IO) error {
	nodeKey, err := readSecretData[p2p.NodeKey](path)
	if err != nil {
		return fmt.Errorf("unable to read node key, %w", err)
	}

	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Node P2P Info]\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Node ID:\t%s\n",
		nodeKey.ID(),
	); err != nil {
		return err
	}

	return w.Flush()
}
