package privval

import (
	"fmt"
	"path/filepath"
	"time"

	rsclient "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// PrivValidatorConfig defines the configuration for the PrivValidator, with a local or remote
// signer, including network parameters and filepaths.
type PrivValidatorConfig struct {
	// Secret path configuration.
	RootDir string `json:"home" toml:"home"`

	// Sign State configuration.
	SignState string `json:"sign_state" toml:"sign_state" comment:"Path to the JSON file containing the last validator state to prevent double-signing"`

	// Local Signer configuration.
	LocalSigner string `json:"local_signer" toml:"local_signer" comment:"Path to the JSON file containing the private key to use for signing using a local signer"`

	// Remote Signer Client configuration.
	RemoteSignerAddress     string        `json:"remote_signer_address" toml:"remote_signer_address" comment:"Address of the remote signer to dial (UNIX or TCP). If set, the local signer is disabled"`
	RemoteAuthorizedKeys    []string      `json:"remote_authorized_keys" toml:"remote_authorized_keys" comment:"List of authorized public keys for the remote signer"`
	RemoteDialMaxRetries    int           `json:"remote_dial_max_retries" toml:"remote_dial_max_retries" comment:"Maximum number of retries to dial the remote signer"`
	RemoteDialRetryInterval time.Duration `json:"remote_dial_retry_interval" toml:"remote_dial_retry_interval" comment:"Interval between retries to dial the remote signer"`
	RemoteDialTimeout       time.Duration `json:"remote_dial_timeout" toml:"remote_dial_timeout" comment:"Timeout to dial the remote signer"`
	RemoteKeepAlivePeriod   time.Duration `json:"remote_keep_alive_period" toml:"remote_keep_alive_period" comment:"Keep alive period for the remote signer connection (TCP only)"`
	RemoteRequestTimeout    time.Duration `json:"remote_request_timeout" toml:"remote_request_timeout" comment:"Timeout for requests to the remote signer"`
}

// PrivValidatorConfig validation errors.
var (
	errInvalidSignStatePath   = errors.New("invalid private validator sign state file path")
	errInvalidLocalSignerPath = errors.New("invalid private validator local signer file path")
	errInvalidAuthorizedKey   = errors.New("invalid private validator remote signer authorized key")
)

// DefaultPrivValidatorConfig returns a default configuration for the PrivValidator.
func DefaultPrivValidatorConfig() *PrivValidatorConfig {
	return &PrivValidatorConfig{
		SignState:               "priv_validator_state.json",
		LocalSigner:             "priv_validator_key.json",
		RemoteSignerAddress:     "", // Empty to disable remote signer by default.
		RemoteDialMaxRetries:    rsclient.DefaultDialMaxRetries,
		RemoteDialRetryInterval: rsclient.DefaultDialRetryInterval,
		RemoteDialTimeout:       rsclient.DefaultDialTimeout,
		RemoteKeepAlivePeriod:   rsclient.DefaultKeepAlivePeriod,
		RemoteRequestTimeout:    rsclient.DefaultRequestTimeout,
		RemoteAuthorizedKeys:    []string{}, // Empty to authorize all keys by default.
	}
}

// TestPrivValidatorConfig returns a configuration for testing the PrivValidator.
func TestPrivValidatorConfig() *PrivValidatorConfig {
	return DefaultPrivValidatorConfig()
}

// SignStatePath returns the complete path for the sign state file.
func (cfg *PrivValidatorConfig) SignStatePath() string {
	return filepath.Join(cfg.RootDir, cfg.SignState)
}

// LocalSignerPath returns the complete path for the local signer file.
func (cfg *PrivValidatorConfig) LocalSignerPath() string {
	return filepath.Join(cfg.RootDir, cfg.LocalSigner)
}

// AuthorizedKeys returns the authorized public keys for the remote signer in ed25519
// format and returns an error if any key is invalid.
func (cfg *PrivValidatorConfig) AuthorizedKeys() ([]ed25519.PubKeyEd25519, error) {
	keys := make([]ed25519.PubKeyEd25519, len(cfg.RemoteAuthorizedKeys))

	for i := range cfg.RemoteAuthorizedKeys {
		// Decode the public key from the Bech32 format.
		pubKey, err := crypto.PubKeyFromBech32(cfg.RemoteAuthorizedKeys[i])
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errInvalidAuthorizedKey, err)
		}

		// Cast the public key to the ed25519 type.
		switch pubKey := pubKey.(type) {
		case ed25519.PubKeyEd25519:
			keys[i] = pubKey

		default:
			return nil, fmt.Errorf("%w: not an ed25519 public key", errInvalidAuthorizedKey)
		}
	}

	return keys, nil
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *PrivValidatorConfig) ValidateBasic() error {
	// Verify the validator sign state file path is set.
	if cfg.SignState == "" {
		return errInvalidSignStatePath
	}

	// Verify the validator local signer file path is set.
	if cfg.LocalSigner == "" {
		return errInvalidLocalSignerPath
	}

	// Verify the authorized keys are valid.
	if _, err := cfg.AuthorizedKeys(); err != nil {
		return err
	}

	return nil
}
