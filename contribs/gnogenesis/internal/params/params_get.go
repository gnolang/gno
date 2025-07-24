package params

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var errInvalidParamsGetArgs = errors.New("invalid number of params get arguments provided")

// newTxsAddPackagesCmd creates the genesis txs add packages subcommand
func newParamsGetCmd(paramsCfg *paramsCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "get",
			ShortUsage: "params get <key>",
			ShortHelp:  "show the Gno genesis params fields value",
			LongHelp:   "Shows the Gno genesis params value by fetching the option specified at <key>",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execParamsGet(paramsCfg, io, args)
		},
	)

	// Add subcommand helpers
	gen := commands.FieldsGenerator{
		MetaUpdate: func(meta *commands.Metadata, inputType string) {
			meta.ShortUsage = fmt.Sprintf("params get %s <%s>", meta.Name, inputType)
		},
		TreeDisplay:     true,
		TagNameSelector: "json",
	}

	cmd.AddSubCommands(gen.GenerateFrom(params{}, func(_ context.Context, args []string) error {
		return execParamsGet(paramsCfg, io, args)
	})...)

	return cmd
}

func execParamsGet(cfg *paramsCfg, io commands.IO, args []string) error {
	genesis, loadErr := types.GenesisDocFromFile(cfg.GenesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Make sure the edit arguments are valid
	if len(args) > 1 {
		return errInvalidParamsGetArgs
	}

	appstate, ok := genesis.AppState.(gnoland.GnoGenesisState)
	if !ok {
		return errInvalidGenesisStateType
	}

	params := params{
		Auth: &appstate.Auth.Params,
		VM:   &appstate.VM.Params,
		Bank: &appstate.Bank.Params,
	}

	if err := printKeyValue(&params, io, args...); err != nil {
		return fmt.Errorf("unable to printout value: %w", err)
	}

	return nil
}

// printKeyValue searches and prints the given key value in JSON
func printKeyValue(input *params, io commands.IO, key ...string) error {
	// prepareOutput prepares the JSON output, taking into account raw mode
	prepareOutput := func(input any) (string, error) {
		encoded, err := json.MarshalIndent(input, "", "    ")
		if err != nil {
			return "", fmt.Errorf("unable to marshal JSON, %w", err)
		}

		return string(encoded), nil
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
	val := reflect.ValueOf(input).Elem()

	// Get the value path, with sections separated out by a period
	field, err := commands.GetFieldByPath(val, "json", strings.Split(key[0], "."))
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
