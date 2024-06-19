package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var (
	errInvalidDataDir    = errors.New("invalid data directory provided")
	errInvalidSecretsKey = errors.New("invalid number of secret key arguments")
)

const (
	defaultValidatorKeyName   = "priv_validator_key.json"
	defaultNodeKeyName        = "node_key.json"
	defaultValidatorStateName = "priv_validator_state.json"
)

const (
	nodeIDKey              = "ID"
	validatorPrivateKeyKey = "ValidatorPrivateKey"
	validatorStateKey      = "ValidatorState"
)

// newSecretsCmd creates the secrets root command
func newSecretsCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "secrets",
			ShortUsage: "secrets <subcommand> [flags] [<arg>...]",
			ShortHelp:  "gno secrets manipulation suite",
			LongHelp:   "gno secrets manipulation suite, for managing the validator key, p2p key and validator state",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newSecretsInitCmd(io),
		newSecretsVerifyCmd(io),
		newSecretsGetCmd(io),
	)

	return cmd
}

// commonAllCfg is the common
// configuration for secrets commands
// that require a bundled secrets dir
type commonAllCfg struct {
	dataDir string
}

func (c *commonAllCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.dataDir,
		"data-dir",
		constructSecretsPath(defaultNodeDir),
		"the secrets output directory",
	)
}

// constructSecretsPath constructs the default secrets path, using
// the given node directory
func constructSecretsPath(nodeDir string) string {
	return filepath.Join(
		nodeDir,
		config.DefaultSecretsDir,
	)
}

// printSecretsValue prints the value of the secret field at the given path
func printSecretsValue(secrets *secrets, key string, io commands.IO) error {
	// Get the secret value using reflect
	secretValue := reflect.ValueOf(secrets).Elem()

	// Get the value path, with sections separated out by a period
	path := strings.Split(key, ".")

	field, err := getFieldAtPath(secretValue, path)
	if err != nil {
		return err
	}

	return outputJSONCommon(field.Interface(), io)
}

// outputJSONCommon outputs the given input to JSON
func outputJSONCommon(input any, io commands.IO) error {
	encoded, err := json.MarshalIndent(input, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to marshal JSON, %w", err)
	}

	io.Println(string(encoded))

	return nil
}

type (
	secrets struct {
		ValidatorKeyInfo   *validatorKeyInfo   `json:"validator_key_info,omitempty" toml:"validator_key_info"`
		ValidatorStateInfo *validatorStateInfo `json:"validator_state_info,omitempty" toml:"validator_state_info"`
		NodeIDInfo         *nodeIDInfo         `json:"node_id_info,omitempty" toml:"node_id_info"`
	}

	validatorKeyInfo struct {
		Address string `json:"address" toml:"address"`
		PubKey  string `json:"pub_key" toml:"pub_key"`
	}

	validatorStateInfo struct {
		Height int64 `json:"height" toml:"height"`
		Round  int   `json:"round" toml:"round"`
		Step   int8  `json:"step" toml:"step"`

		Signature []byte `json:"signature,omitempty" toml:"signature,omitempty"`
		SignBytes []byte `json:"sign_bytes,omitempty" toml:"sign_bytes,omitempty"`
	}

	nodeIDInfo struct {
		ID         string `json:"id" json:"id"`
		P2PAddress string `json:"p2p_address" toml:"p2p_address"`
	}
)
