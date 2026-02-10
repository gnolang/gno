package params

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errInvalidParamsSetArgs = errors.New("invalid number of params set arguments provided")

// newTxsAddPackagesCmd creates the genesis txs add packages subcommand
func newParamsSetCmd(paramsCfg *paramsCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "set",
			ShortUsage: "params set <key> <...values>",
			ShortHelp:  "edit the Gno genesis params fields value",
			LongHelp: "Edits params configuration of the given genesis path " +
				"by setting the option specified at <key> to the given <values>",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execParamsSet(paramsCfg, io, args)
		},
	)

	// Add subcommand helpers
	gen := commands.FieldsGenerator{
		MetaUpdate: func(meta *commands.Metadata, inputType string) {
			meta.ShortUsage = fmt.Sprintf(" set %s <%s>", meta.Name, inputType)
		},
		TagNameSelector: "json",
		Depth:           1,
	}

	cmd.AddSubCommands(gen.GenerateFrom(params{}, func(_ context.Context, args []string) error {
		return execParamsSet(paramsCfg, io, args)
	})...)

	return cmd
}

func execParamsSet(cfg *paramsCfg, io commands.IO, args []string) error {
	genesis, loadErr := types.GenesisDocFromFile(cfg.GenesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Make sure the edit arguments are valid
	if len(args) < 2 {
		return errInvalidParamsSetArgs
	}

	key, vals := args[0], args[1:]
	appstate := genesis.AppState.(gnoland.GnoGenesisState)
	params := params{
		Auth: &appstate.Auth.Params,
		VM:   &appstate.VM.Params,
		Bank: &appstate.Bank.Params,
	}

	err := updateParamsField(&params, key, vals)
	if err != nil {
		return fmt.Errorf("unable to set params %q: %w", key, err)
	}

	// Override AppState with the updated one
	genesis.AppState = appstate

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.GenesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln("Saved state to %q", cfg.GenesisPath)

	return nil
}

// updateFieldAtPath updates the field at the given path, with the given value
func updateParamsField(prms *params, key string, vals []string) error {
	// Get the config value using reflect
	configValue := reflect.ValueOf(prms).Elem()

	// Get the value path, with sections separated out by a period
	path := strings.Split(key, ".")

	// Get the editable field
	field, err := commands.GetFieldByPath(configValue, "json", path)
	if err != nil {
		return err
	}

	// Attempt to update the field value
	if err = saveStringToValue(vals, *field); err != nil {
		return fmt.Errorf("unable to convert value to field type %q, %w", field.Type().String(), err)
	}

	return nil
}

// saveStringToValue attempts to convert the given
// string value to the destination type and save it to the destination value.
// Because we opted to using reflect instead of a flag-based approach,
// arguments (always strings) need to be converted to the field's
// respective type, if possible
func saveStringToValue(vals []string, dstValue reflect.Value) error {
	if len(vals) == 0 {
		return fmt.Errorf("no value(s) to set")
	}

	switch dstValue.Interface().(type) {
	case string:
		dstValue.Set(reflect.ValueOf(vals[0]))
	case []string:
		res := []string{}
		for _, val := range vals {
			if val == "" {
				continue
			}

			res = append(res, strings.Split(val, ",")...)
		}

		dstValue.Set(reflect.ValueOf(res))
	case crypto.Address:
		addr, err := crypto.AddressFromBech32(vals[0])
		if err != nil {
			return fmt.Errorf("unable to parse address %q: %w", vals[0], err)
		}
		dstValue.Set(reflect.ValueOf(addr))

	case []crypto.Address:
		addrs := []crypto.Address{}
		for _, val := range vals {
			if val == "" {
				continue
			}

			for _, ss := range strings.Split(val, ",") {
				addr, err := crypto.AddressFromBech32(ss)
				if err != nil {
					return fmt.Errorf("unable to parse address %q: %w", ss, err)
				}
				addrs = append(addrs, addr)
			}
		}
		dstValue.Set(reflect.ValueOf(addrs))

	case std.GasPrice:
		gas, err := std.ParseGasPrice(vals[0])
		if err != nil {
			return fmt.Errorf("unable to parse gas price %q: %w", vals[0], err)
		}

		dstValue.Set(reflect.ValueOf(gas))
	case time.Duration:
		d, err := time.ParseDuration(vals[0])
		if err != nil {
			return fmt.Errorf("unable to parse time.Duration, %w", err)
		}

		dstValue.Set(reflect.ValueOf(d))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return json.Unmarshal([]byte(vals[0]), dstValue.Addr().Interface())
	default:
		return fmt.Errorf("unsupported type, %s", dstValue.Type().Name())
	}

	return nil
}
