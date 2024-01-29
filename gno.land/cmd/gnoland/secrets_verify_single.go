package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type verifySingleCfg struct {
	commonSingleCfg
}

// newVerifySingleCmd creates the secrets verify single command
func newVerifySingleCmd(io commands.IO) *commands.Command {
	cfg := &verifySingleCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "single",
			ShortUsage: "secrets verify single [flags]",
			ShortHelp:  "Verifies required Gno secrets individually",
			LongHelp: "Verifies the validator private key, the node p2p key and the validator's last sign state" +
				" at custom paths",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execVerifySingle(cfg, io)
		},
	)

	return cmd
}

func (c *verifySingleCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonSingleCfg.RegisterFlags(fs)
}

func execVerifySingle(cfg *verifySingleCfg, io commands.IO) error {
	var (
		validatorKeyPathSet   = cfg.validatorKeyPath != ""
		validatorStatePathSet = cfg.validatorStatePath != ""
		nodeKeyPathSet        = cfg.nodeKeyPath != ""
	)

	var (
		validatorKey   *privval.FilePVKey
		validatorState *privval.FilePVLastSignState

		err error
	)

	if !validatorKeyPathSet && !validatorStatePathSet && !nodeKeyPathSet {
		return errNoOutputSet
	}

	// Verify the validator private key, if any
	if validatorKeyPathSet {
		validatorKey, err = readAndVerifyValidatorKey(cfg.validatorKeyPath, io)
		if err != nil {
			return err
		}
	}

	// Verify the last validator sign state, if any
	if validatorStatePathSet {
		validatorState, err = readAndVerifyValidatorState(cfg.validatorStatePath, io)
		if err != nil {
			return err
		}
	}

	// Verify the signature bytes if the key and validator state
	// is provided
	if validatorKey != nil && validatorState != nil {
		if err = validateValidatorStateSignature(validatorState, validatorKey.PubKey); err != nil {
			return err
		}
	}

	// Verify the node key, if any
	if nodeKeyPathSet {
		if err = readAndVerifyNodeKey(cfg.nodeKeyPath, io); err != nil {
			return err
		}
	}

	return nil
}
