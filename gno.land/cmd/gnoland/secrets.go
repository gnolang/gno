package main

import (
	"errors"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidDataDir = errors.New("invalid data directory provided")

const (
	defaultSecretsDir         = "./secrets"
	defaultValidatorKeyName   = "priv_validator_key.json"
	defaultNodeKeyName        = "node_key.json"
	defaultValidatorStateName = "priv_validator_state.json"
)

const (
	nodeKeyKey             = "NodeKey"
	validatorPrivateKeyKey = "ValidatorPrivateKey"
	validatorStateKey      = "ValidatorStateKey"
)

// newSecretsCmd creates the secrets root command
func newSecretsCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "secrets",
			ShortUsage: "gnoland secrets <subcommand> [flags] [<arg>...]",
			ShortHelp:  "gno secrets manipulation suite",
			LongHelp:   "gno secrets manipulation suite, for managing the validator key, p2p key and validator state",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newSecretsInitCmd(io),
		newSecretsVerifyCmd(io),
		newSecretsGetCmd(io),
	)

	return cmd
}

// commonAllCfg is the common
// configuration for secrets commands
// that require a bundled secrets dir
type commonAllCfg struct {
	dataDir string
}

func (c *commonAllCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.dataDir,
		"data-dir",
		defaultSecretsDir,
		"the secrets output directory",
	)
}

// verifySecretsKey verifies the secrets key value from the passed in arguments
func verifySecretsKey(args []string) error {
	if len(args) == 0 {
		return nil
	}

	if len(args) > 1 {
		return errors.New("invalid number of secret key arguments")
	}

	key := args[0]

	if key != nodeKeyKey &&
		key != validatorPrivateKeyKey &&
		key != validatorStateKey {
		return errors.New("invalid secrets key value")
	}

	return nil
}
