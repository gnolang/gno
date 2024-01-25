package main

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidDataDir = errors.New("invalid data directory provided")

const (
	defaultSecretsDir         = "./secrets"
	defaultValidatorKeyName   = "priv_validator_key.json"
	defaultNodeKeyName        = "node_key.json"
	defaultValidatorStateName = "priv_validator_state.json"
)

func main() {
	io := commands.NewDefaultIO()
	cmd := newRootCmd(io)

	cmd.Execute(context.Background(), os.Args[1:])
}

// newRootCmd creates the new secrets root command
func newRootCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Gno secrets manipulation suite",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newInitCmd(io),
		// newVerifyCmd(io),
		// newShowCmd(io),
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

// commonIndividualCfg is the common
// configuration for secrets commands
// that require individual secret path management
type commonIndividualCfg struct {
	validatorKeyPath   string
	validatorStatePath string
	nodeKeyPath        string
}

func (c *commonIndividualCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.validatorKeyPath,
		"validator-key-path",
		"",
		"the path to the validator private key",
	)

	fs.StringVar(
		&c.validatorStatePath,
		"validator-state-path",
		"",
		"the path to the last validator state",
	)

	fs.StringVar(
		&c.nodeKeyPath,
		"node-key-path",
		"",
		"the path to the node p2p key",
	)
}
