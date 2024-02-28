package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidConfigOutputPath = errors.New("invalid config output path provided")

type configInitCfg struct {
	outputPath string
}

// newConfigInitCmd creates the config init command
func newConfigInitCmd(io commands.IO) *commands.Command {
	cfg := &configInitCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "config init [flags]",
			ShortHelp:  "Initializes the Gno node configuration",
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
	fs.StringVar(
		&c.outputPath,
		"output-path",
		"./config.toml",
		"the output path for the config.toml",
	)
}

func execConfigInit(cfg *configInitCfg, io commands.IO) error {
	// Check the config output path
	if cfg.outputPath == "" {
		return errInvalidConfigOutputPath
	}

	// Get the default config
	c := config.DefaultConfig()

	// Save the config to the path
	if err := config.WriteConfigFile(cfg.outputPath, c); err != nil {
		return fmt.Errorf("unable to initialize config, %w", err)
	}

	io.Printfln("Default configuration initialized at %s", cfg.outputPath)

	return nil
}
