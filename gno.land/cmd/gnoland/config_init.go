package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

type configInitCfg struct {
	rootCfg

	forceOverwrite bool
}

// newConfigInitCmd creates the config init command
func newConfigInitCmd(io commands.IO) *commands.Command {
	cfg := &configInitCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "config init [flags]",
			ShortHelp:  "initializes the Gno node configuration",
			LongHelp: "Initializes the Gno node configuration locally with default values, which includes" +
				" the base and module configurations",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execConfigInit(cfg, io)
		},
	)

	return cmd
}

func (c *configInitCfg) RegisterFlags(fs *flag.FlagSet) {
	c.rootCfg.RegisterFlags(fs)

	fs.BoolVar(
		&c.forceOverwrite,
		"force",
		false,
		"overwrite existing config.toml, if any",
	)
}

func execConfigInit(cfg *configInitCfg, io commands.IO) error {
	// Make sure overwriting the config is enabled
	if osm.FileExists(cfg.homeDir.ConfigFile()) && !cfg.forceOverwrite {
		return errOverwriteNotEnabled
	}

	// Get the default config
	c := config.DefaultConfig()

	// Make sure the path is created
	if err := os.MkdirAll(cfg.homeDir.ConfigDir(), 0o755); err != nil {
		return fmt.Errorf("unable to create config dir, %w", err)
	}

	// Save the config to the path
	if err := config.WriteConfigFile(cfg.homeDir.ConfigFile(), c); err != nil {
		return fmt.Errorf("unable to initialize config, %w", err)
	}

	io.Printfln("Default configuration initialized at %s", cfg.homeDir.ConfigFile())

	return nil
}
