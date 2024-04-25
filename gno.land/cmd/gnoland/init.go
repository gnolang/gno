package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type initCfg struct {
	dataDir        string
	forceOverwrite bool
}

// newInitCmd creates the gnoland init command
func newInitCmd(io commands.IO) *commands.Command {
	cfg := &initCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [flags]",
			ShortHelp:  "initializes the default node secrets / configuration",
			LongHelp:   "initializes the node directory containing the secrets and configuration files",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execInit(cfg, io)
		},
	)
}

func (c *initCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.dataDir,
		"data-dir",
		defaultNodeDir,
		"the path to the node's data directory",
	)

	fs.BoolVar(
		&c.forceOverwrite,
		"force",
		false,
		"overwrite existing data, if any",
	)
}

func execInit(cfg *initCfg, io commands.IO) error {
	// Create the gnoland config options
	config := &configInitCfg{
		configCfg: configCfg{
			configPath: constructConfigPath(cfg.dataDir),
		},
		forceOverwrite: cfg.forceOverwrite,
	}

	// Run gnoland config init
	if err := execConfigInit(config, io); err != nil {
		return fmt.Errorf("unable to initialize config, %w", err)
	}

	// Create the gnoland secrets options
	secrets := &secretsInitCfg{
		commonAllCfg: commonAllCfg{
			dataDir: constructSecretsPath(cfg.dataDir),
		},
		forceOverwrite: cfg.forceOverwrite,
	}

	// Run gnoland secrets init
	if err := execSecretsInit(secrets, []string{}, io); err != nil {
		return fmt.Errorf("unable to initialize secrets, %w", err)
	}

	io.Println()

	io.Printfln("Successfully initialized default node config at %q", filepath.Dir(config.configPath))
	io.Printfln("Successfully initialized default node secrets at %q", secrets.dataDir)

	return nil
}
