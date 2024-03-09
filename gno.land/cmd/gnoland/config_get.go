package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidConfigGetArgs = errors.New("invalid number of config get arguments provided")

// newConfigGetCmd creates the config get command
func newConfigGetCmd(io commands.IO) *commands.Command {
	cfg := &configCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "get",
			ShortUsage: "config get <key>",
			ShortHelp:  "shows the Gno node configuration",
			LongHelp: "Shows the Gno node configuration at the given path " +
				"by fetching the option specified at <key>",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execConfigGet(cfg, io, args)
		},
	)

	return cmd
}

func execConfigGet(cfg *configCfg, io commands.IO, args []string) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("unable to load config, %w", err)
	}

	// Make sure the edit arguments are valid
	if len(args) != 1 {
		return errInvalidConfigGetArgs
	}

	// Find and print the config field, if any
	if err := printConfigField(loadedCfg, args[0], io); err != nil {
		return fmt.Errorf("unable to update config field, %w", err)
	}

	return nil
}

// printConfigField prints the value of the field at the given path
func printConfigField(config *config.Config, key string, io commands.IO) error {
	// Get the config value using reflect
	configValue := reflect.ValueOf(config).Elem()

	// Get the value path, with sections separated out by a period
	path := strings.Split(key, ".")

	field, err := getFieldAtPath(configValue, path)
	if err != nil {
		return err
	}

	io.Printf("%v", field.Interface())

	return nil
}
