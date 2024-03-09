package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidConfigSetArgs = errors.New("invalid number of config set arguments provided")

// newConfigSetCmd creates the config set command
func newConfigSetCmd(io commands.IO) *commands.Command {
	cfg := &configCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "set",
			ShortUsage: "config set <key> <value>",
			ShortHelp:  "Edits the Gno node configuration",
			LongHelp: "Edits the Gno node configuration at the given path " +
				"by setting the option specified at <key> to the given <value>",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execConfigEdit(cfg, io, args)
		},
	)

	return cmd
}

func execConfigEdit(cfg *configCfg, io commands.IO, args []string) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("unable to load config, %w", err)
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

	field, err := getFieldAtPath(configValue, path)
	if err != nil {
		return err
	}

	// We've reached the actual field, check if it can be updated
	if !field.CanSet() {
		return fmt.Errorf("unable to set value")
	}

	// Convert the value to the field's type
	v, err := convertStringToType(value, field.Interface())
	if err != nil {
		return fmt.Errorf("unable to convert value to field type, %w", err)
	}

	// Update the field value
	field.Set(reflect.ValueOf(v))

	return nil
}

// generateInvalidFieldError generates an invalid field error
func generateInvalidFieldError(field string, value reflect.Value) error {
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
		"field \"%s\", is not a valid configuration key, available keys: %s",
		field,
		fields,
	)
}

// convertStringToType attempts to convert the given
// string value to an output type.
// Because we opted to using reflect instead of a flag-based approach,
// arguments (always strings) need to be converted to the field's
// respective type, if possible
func convertStringToType(value string, outputType any) (any, error) {
	parseInt := func(size int) (any, error) {
		castValue, err := strconv.ParseInt(value, 10, size)
		if err != nil {
			return nil, fmt.Errorf("unable to convert to int%d, %w", size, err)
		}

		return castValue, nil
	}

	parseUint := func(size int) (any, error) {
		castValue, err := strconv.ParseInt(value, 10, size)
		if err != nil {
			return nil, fmt.Errorf("unable to convert to uint%d, %w", size, err)
		}

		return castValue, nil
	}

	parseFloat := func(size int) (any, error) {
		castValue, err := strconv.ParseFloat(value, size)
		if err != nil {
			return nil, fmt.Errorf("unable to convert to float%d, %w", size, err)
		}

		return castValue, nil
	}

	switch outputType.(type) {
	case string:
		return value, nil
	case []string:
		// This is a special case.
		// Since values are given as a single string (argument),
		// they need to be parsed from a custom format.
		// In this case, the format for a []string is comma separated:
		// value1,value2,value3 ...
		return strings.SplitN(value, ",", -1), nil
	case time.Duration:
		castValue, err := time.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("unable to convert to Duration, %w", err)
		}

		return castValue, nil
	case int:
		castValue, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("unable to convert to int, %w", err)
		}

		return castValue, nil
	case int8:
		return parseInt(8)
	case int16:
		return parseInt(16)
	case int32:
		return parseInt(32)
	case int64:
		return parseInt(64)
	case bool:
		castValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("unable to convert to bool, %w", err)
		}

		return castValue, nil
	case uint, uint64:
		return parseUint(64)
	case uint8:
		return parseUint(8)
	case uint16:
		return parseUint(16)
	case uint32:
		return parseUint(32)
	case float32:
		return parseFloat(32)
	case float64:
		return parseFloat(64)
	case types.EventStoreParams:
		// This is a special case.
		// Map values are tricky to parse, especially
		// since it's a custom type alias, so this
		// method is used to parse out the key value pairs
		// that are given in a custom format,
		// for the event store params
		return parseEventStoreParams(value), nil
	default:
		return nil, fmt.Errorf("unsupported type, %s", outputType)
	}
}

// parseEventStoreParams parses the event store params into a param map.
// Map values are provided in the format <key>=<value> and comma separated
// for different keys: <key1>=<value1>,<key2>=<value2>
func parseEventStoreParams(values string) types.EventStoreParams {
	params := make(types.EventStoreParams, len(values))

	// Split the string into different key value pairs
	keyPairs := strings.SplitN(values, ",", -1)

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
