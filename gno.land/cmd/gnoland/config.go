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
	field, err := commands.GetFieldByPath(secretValue, "toml", strings.Split(key[0], "."))
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
