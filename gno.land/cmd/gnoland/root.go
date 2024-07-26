package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
)

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}

type rootCfg struct {
	homeDir homeDirectory
}

func (c *rootCfg) RegisterFlags(fs *flag.FlagSet) {
	if home := fs.Lookup("home"); home == nil {
		fs.StringVar(
			&c.homeDir.homeDir,
			"home",
			defaultNodeDir,
			"Directory for config, secrets and data",
		)
	} else {
		c.homeDir.homeDir = home.Value.(flag.Getter).Get().(string)
	}
}

func newRootCmd(io commands.IO) *commands.Command {
	cfg := &rootCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "starts the gnoland blockchain node",
			Options: []ff.Option{
				ff.WithConfigFileParser(fftoml.Parser),
			},
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newStartCmd(io),
		newGenesisCmd(io),
		newSecretsCmd(io),
		newConfigCmd(io),
	)

	return cmd
}

type homeDirectory struct {
	homeDir     string
	genesisFile string
}

func (h homeDirectory) Path() string       { return h.homeDir }
func (h homeDirectory) ConfigDir() string  { return filepath.Join(h.Path(), "/config") }
func (h homeDirectory) ConfigFile() string { return filepath.Join(h.ConfigDir(), "/config.toml") }

func (h homeDirectory) GenesisFilePath() string {
	if h.genesisFile != "" {
		return h.genesisFile
	}
	return filepath.Join(h.Path(), "/genesis.json")
}

func (h homeDirectory) SecretsDir() string { return filepath.Join(h.Path(), "/secrets") }
func (h homeDirectory) SecretsNodeKey() string {
	return filepath.Join(h.SecretsDir(), defaultNodeKeyName)
}

func (h homeDirectory) SecretsValidatorKey() string {
	return filepath.Join(h.SecretsDir(), defaultValidatorKeyName)
}

func (h homeDirectory) SecretsValidatorState() string {
	return filepath.Join(h.SecretsDir(), defaultValidatorStateName)
}

func (h homeDirectory) GetSecrets() (*secrets, error) {
	return loadSecrets(h)
}

func (h homeDirectory) GetConfig() (*config.Config, error) {
	return config.LoadConfig(h.Path())
}
