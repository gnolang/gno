package main

import (
	"errors"
	"flag"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
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
	nodeIDKey              = "node_id"
	validatorPrivateKeyKey = "validator_key"
	validatorStateKey      = "validator_state"
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
		ValidatorKeyInfo   *validatorKeyInfo   `json:"validator_key,omitempty" toml:"validator_key" comment:"the validator private key info"`
		ValidatorStateInfo *validatorStateInfo `json:"validator_state,omitempty" toml:"validator_state" comment:"the last signed validator state info"`
		NodeIDInfo         *nodeIDInfo         `json:"node_id,omitempty" toml:"node_id" comment:"the derived node ID info"`
	}

	// NOTE: keep in sync with tm2/pkg/bft/privval/state/state.go
	validatorKeyInfo struct {
		Address string `json:"address" toml:"address" comment:"the validator address"`
		PubKey  string `json:"pub_key" toml:"pub_key" comment:"the validator public key"`
	}

	// NOTE: keep in sync with tm2/pkg/bft/privval/signer/local/key.go
	validatorStateInfo struct {
		Height int64       `json:"height" toml:"height" comment:"the height of the last sign"`
		Round  int         `json:"round" toml:"round" comment:"the round of the last sign"`
		Step   fstate.Step `json:"step" toml:"step" comment:"the step of the last sign"`

		Signature []byte `json:"signature,omitempty" toml:"signature,omitempty" comment:"the signature of the last sign"`
		SignBytes []byte `json:"sign_bytes,omitempty" toml:"sign_bytes,omitempty" comment:"the raw signature bytes of the last sign"`
	}

	// NOTE: keep in sync with tm2/pkg/p2p/types/key.go
	nodeIDInfo struct {
		ID         string `json:"id" toml:"id" comment:"the node ID derived from the private key"`
		P2PAddress string `json:"p2p_address" toml:"p2p_address" comment:"the node's constructed P2P address'"`
		PubKey     string `json:"pub_key" toml:"pub_key" comment:"the node public key that can be used to anthenticate with gnokms"`
	}
)
