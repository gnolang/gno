package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

const tryConfigInit = "unable to load config; try running `gnoland config init` or use the -lazy flag"

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
		constructConfigPath(defaultNodeDir),
		"the path for the config.toml",
	)
}

// constructConfigPath constructs the default config path, using
// the given node directory
func constructConfigPath(nodeDir string) string {
	return filepath.Join(
		nodeDir,
		config.DefaultConfigDir,
		config.DefaultConfigFileName,
	)
}

// printKeyValue searches and prints the given key value in JSON
func printKeyValue[T *secrets | *config.Config](
	input T,
	raw bool,
	io commands.IO,
	key ...string,
) error {
	// prepareOutput prepares the JSON output, taking into account raw mode
	prepareOutput := func(input any) (string, error) {
		encoded, err := json.MarshalIndent(input, "", "    ")
		if err != nil {
			return "", fmt.Errorf("unable to marshal JSON, %w", err)
		}

		output := string(encoded)

		if raw {
			if err := json.Unmarshal(encoded, &output); err != nil {
				return "", fmt.Errorf("unable to unmarshal raw JSON, %w", err)
			}
		}

		return output, nil
	}

	if len(key) == 0 {
		// Print the entire input
		output, err := prepareOutput(input)
		if err != nil {
			return err
		}

		io.Println(output)

		return nil
	}

	// Get the value using reflect
	secretValue := reflect.ValueOf(input).Elem()

	// Get the value path, with sections separated out by a period
	field, err := getFieldAtPath(secretValue, strings.Split(key[0], "."))
	if err != nil {
		return err
	}

	output, err := prepareOutput(field.Interface())
	if err != nil {
		return err
	}

	io.Println(output)

	return nil
}

// getFieldAtPath fetches the given field from the given path
func getFieldAtPath(currentValue reflect.Value, path []string) (*reflect.Value, error) {
	// Look at the current section, and figure out if
	// it's a part of the current struct
	field := fieldByTOMLName(currentValue, path[0])
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

func fieldByTOMLName(value reflect.Value, name string) reflect.Value {
	var dst reflect.Value
	eachTOMLField(value, func(val reflect.Value, tomlName string) bool {
		if tomlName == name {
			dst = val
			return true
		}
		return false
	})
	return dst
}

// eachTOMLField iterates over each field in value (assumed to be a struct).
// For every field within the struct, iterationCallback is called with the value
// of the field and the associated name in the TOML representation
// (through the `toml:"..."` struct field, or just the field's name as fallback).
// If iterationCallback returns true, the function returns immediately with
// true. If it always returns false, or no fields were processed, eachTOMLField
// will return false.
func eachTOMLField(value reflect.Value, iterationCallback func(val reflect.Value, tomlName string) bool) bool {
	// For reference:
	// https://github.com/pelletier/go-toml/blob/7dad87762adb203e30b96a46026d1428ef2491a2/unmarshaler.go#L1251-L1270

	currentType := value.Type()
	nf := currentType.NumField()
	for i := 0; i < nf; i++ {
		fld := currentType.Field(i)
		tomlName := fld.Tag.Get("toml")

		// Ignore `toml:"-"`, strip away any "omitempty" or other options.
		if tomlName == "-" || !fld.IsExported() {
			continue
		}
		if pos := strings.IndexByte(tomlName, ','); pos != -1 {
			tomlName = tomlName[:pos]
		}

		// Handle anonymous (embedded) fields.
		// Anonymous fields will be treated regularly if they have a tag.
		if fld.Anonymous && tomlName == "" {
			anon := fld.Type
			if anon.Kind() == reflect.Ptr {
				anon = anon.Elem()
			}

			if anon.Kind() == reflect.Struct {
				// NOTE: naive, if there is a conflict the embedder should take
				// precedence over the embedded; but the TOML parser seems to
				// ignore this, too, and this "unmarshaler" isn't fit for general
				// purpose.
				if eachTOMLField(value.Field(i), iterationCallback) {
					return true
				}
				continue
			}
			// If it's not a struct, or *struct, it should be treated regularly.
		}

		// general case, simple struct field.
		if tomlName == "" {
			tomlName = fld.Name
		}
		if iterationCallback(value.Field(i), tomlName) {
			return true
		}
	}
	return false
}

// newInvalidFieldError creates an error for non-existent struct fields
// being passed as arguments to [getFieldAtPath]
func newInvalidFieldError(field string, value reflect.Value) error {
	fields := make([]string, 0, value.NumField())

	eachTOMLField(value, func(val reflect.Value, tomlName string) bool {
		fields = append(fields, tomlName)
		return false
	})

	return fmt.Errorf(
		"field %q, is not a valid configuration key, available keys: %s",
		field,
		fields,
	)
}
