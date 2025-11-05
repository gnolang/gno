package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	storeTypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

var errInvalidConfigSetArgs = errors.New("invalid number of config set arguments provided")

// newConfigSetCmd creates the config set command
func newConfigSetCmd(io commands.IO) *commands.Command {
	cfg := &configCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "set",
			ShortUsage: "config set <key> <value>",
			ShortHelp:  "edits the Gno node configuration",
			LongHelp: "Edits the Gno node configuration at the given path " +
				"by setting the option specified at <key> to the given <value>",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execConfigEdit(cfg, io, args)
		},
	)

	// Add subcommand helpers
	gen := commands.FieldsGenerator{
		MetaUpdate: func(meta *commands.Metadata, inputType string) {
			meta.ShortUsage = fmt.Sprintf("config set %s <%s>", meta.Name, inputType)
		},
		TagNameSelector: "json",
		TreeDisplay:     true,
	}

	cmd.AddSubCommands(gen.GenerateFrom(config.Config{}, func(_ context.Context, args []string) error {
		return execConfigEdit(cfg, io, args)
	})...)

	return cmd
}

func execConfigEdit(cfg *configCfg, io commands.IO, args []string) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("%s, %w", tryConfigInit, err)
	}

	// Make sure the edit arguments are valid
	if len(args) != 2 {
		return errInvalidConfigSetArgs
	}

	var (
		key   = args[0]
		value = args[1]
	)

	// Update the config field
	if err := updateConfigField(
		loadedCfg,
		key,
		value,
	); err != nil {
		return fmt.Errorf("unable to update config field, %w", err)
	}

	// Make sure the config is now valid
	if err := loadedCfg.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate config, %w", err)
	}

	// Save the config
	if err := config.WriteConfigFile(cfg.configPath, loadedCfg); err != nil {
		return fmt.Errorf("unable to save updated config, %w", err)
	}

	io.Printfln("Updated configuration saved at %s", cfg.configPath)

	return nil
}

// updateFieldAtPath updates the field at the given path, with the given value
func updateConfigField(config *config.Config, key, value string) error {
	// Get the config value using reflect
	configValue := reflect.ValueOf(config).Elem()

	// Get the value path, with sections separated out by a period
	path := strings.Split(key, ".")

	// Get the editable field
	field, err := commands.GetFieldByPath(configValue, "toml", path)
	if err != nil {
		return err
	}

	// Attempt to update the field value
	if err = saveStringToValue(value, *field); err != nil {
		return fmt.Errorf("unable to convert value to field type, %w", err)
	}

	return nil
}

// saveStringToValue attempts to convert the given
// string value to the destination type and save it to the destination value.
// Because we opted to using reflect instead of a flag-based approach,
// arguments (always strings) need to be converted to the field's
// respective type, if possible
func saveStringToValue(value string, dstValue reflect.Value) error {
	switch dstValue.Interface().(type) {
	case string:
		dstValue.Set(reflect.ValueOf(value))
	case []string:
		// This is a special case.
		// Since values are given as a single string (argument),
		// they need to be parsed from a custom format.
		// In this case, the format for a []string is comma separated:
		// value1,value2,value3 ...
		val := strings.Split(value, ",")

		dstValue.Set(reflect.ValueOf(val))
	case time.Duration:
		val, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("unable to parse time.Duration, %w", err)
		}

		dstValue.Set(reflect.ValueOf(val))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return json.Unmarshal([]byte(value), dstValue.Addr().Interface())
	case types.EventStoreParams:
		// This is a special case.
		// Map values are tricky to parse, especially
		// since it's a custom type alias, so this
		// method is used to parse out the key value pairs
		// that are given in a custom format,
		// for the event store params
		val := parseEventStoreParams(value)

		dstValue.Set(reflect.ValueOf(val))
	case storeTypes.PruneStrategy:
		var (
			dstType = dstValue.Type()
			val     = reflect.ValueOf(value)
		)

		if !val.CanConvert(dstType) {
			return fmt.Errorf("unable to cast to type %q", dstType)
		}

		dstValue.Set(val.Convert(dstType))
	default:
		return fmt.Errorf("unsupported type, %s", dstValue.Type().Name())
	}

	return nil
}

// parseEventStoreParams parses the event store params into a param map.
// Map values are provided in the format <key>=<value> and comma separated
// for different keys: <key1>=<value1>,<key2>=<value2>
func parseEventStoreParams(values string) types.EventStoreParams {
	params := make(types.EventStoreParams, len(values))

	// Split the string into different key value pairs
	keyPairs := strings.Split(values, ",")

	for _, keyPair := range keyPairs {
		// Split the string into key and value
		kv := strings.SplitN(keyPair, "=", 2)

		// Check if the split produced exactly two elements
		if len(kv) != 2 {
			continue
		}

		key := kv[0]
		value := kv[1]

		params[key] = value
	}

	return params
}
