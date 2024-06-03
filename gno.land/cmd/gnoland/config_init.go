package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidConfigOutputPath = errors.New("invalid config output path provided")

// newConfigInitCmd creates the config init command
func newConfigInitCmd(io commands.IO) *commands.Command {
	cfg := &configCfg{}

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

func execConfigInit(cfg *configCfg, io commands.IO) error {
	// Check the config output path
	if cfg.configPath == "" {
		return errInvalidConfigOutputPath
	}

	// Get the default config
	c := config.DefaultConfig()

	// Save the config to the path
	if err := config.WriteConfigFile(cfg.configPath, c); err != nil {
		return fmt.Errorf("unable to initialize config, %w", err)
	}

	io.Printfln("Default configuration initialized at %s", cfg.configPath)

	return nil
}
