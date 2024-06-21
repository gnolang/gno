package main

import (
	"errors"
	"flag"
	"path/filepath"

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
