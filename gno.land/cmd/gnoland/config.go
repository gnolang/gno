package main

import (
	"flag"
	"fmt"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type configCfg struct {
	configPath string
}

// newConfigCmd creates the config root command
func newConfigCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "config",
			ShortUsage: "config <subcommand> [flags]",
			ShortHelp:  "gno config manipulation suite",
			LongHelp:   "Gno config manipulation suite, for editing base and module configurations",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newConfigInitCmd(io),
		newConfigSetCmd(io),
		newConfigGetCmd(io),
	)

	return cmd
}

func (c *configCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.configPath,
		"config-path",
		"./config.toml",
		"the path for the config.toml",
	)
}

// getFieldAtPath fetches the given field from the given path
func getFieldAtPath(currentValue reflect.Value, path []string) (*reflect.Value, error) {
	// Look at the current section, and figure out if
	// it's a part of the current struct
	field := currentValue.FieldByName(path[0])
	if !field.IsValid() || !field.CanSet() {
		return nil, newInvalidFieldError(path[0], currentValue)
	}

	// Dereference the field if needed
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	// Check if this is not the end of the path
	// ex: x.y.field
	if len(path) > 1 {
		// Recursively try to traverse the path and return the given field
		return getFieldAtPath(field, path[1:])
	}

	return &field, nil
}

// newInvalidFieldError creates an error for non-existent struct fields
// being passed as arguments to [getFieldAtPath]
func newInvalidFieldError(field string, value reflect.Value) error {
	var (
		valueType = value.Type()
		numFields = value.NumField()
	)

	fields := make([]string, 0, numFields)

	for i := 0; i < numFields; i++ {
		valueField := valueType.Field(i)
		if !valueField.IsExported() {
			continue
		}

		fields = append(fields, valueField.Name)
	}

	return fmt.Errorf(
		"field %q, is not a valid configuration key, available keys: %s",
		field,
		fields,
	)
}
