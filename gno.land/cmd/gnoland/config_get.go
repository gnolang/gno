package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidConfigGetArgs = errors.New("invalid number of config get arguments provided")

type configGetCfg struct {
	configCfg

	raw bool
}

// newConfigGetCmd creates the config get command
func newConfigGetCmd(io commands.IO) *commands.Command {
	cfg := &configGetCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "get",
			ShortUsage: "config get [flags] [<key>]",
			ShortHelp:  "shows the Gno node configuration",
			LongHelp: "Shows the Gno node configuration at the given path " +
				"by fetching the option specified at <key>",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execConfigGet(cfg, io, args)
		},
	)

	// Add subcommand helpers
	gen := commands.FieldsGenerator{
		MetaUpdate: func(meta *commands.Metadata, inputType string) {
			meta.ShortUsage = fmt.Sprintf("config get %s <%s>", meta.Name, inputType)
		},
		TagNameSelector: "json",
		TreeDisplay:     true,
	}

	cmd.AddSubCommands(gen.GenerateFrom(config.Config{}, func(_ context.Context, args []string) error {
		return execConfigGet(cfg, io, args)
	})...)

	return cmd
}

func (c *configGetCfg) RegisterFlags(fs *flag.FlagSet) {
	c.configCfg.RegisterFlags(fs)

	fs.BoolVar(
		&c.raw,
		"raw",
		false,
		"output raw string values, rather than as JSON strings",
	)
}

func execConfigGet(cfg *configGetCfg, io commands.IO, args []string) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("%s, %w", tryConfigInit, err)
	}

	// Make sure the get arguments are valid
	if len(args) > 1 {
		return errInvalidConfigGetArgs
	}

	// Find and print the config field, if any
	if err := printKeyValue(loadedCfg, cfg.raw, io, args...); err != nil {
		return fmt.Errorf("unable to get config field, %w", err)
	}

	return nil
}
