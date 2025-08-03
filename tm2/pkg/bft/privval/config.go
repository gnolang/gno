package privval

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	rsclient "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/client"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// PrivValidatorConfig defines the configuration for the PrivValidator, with a local or remote
// signer, including network parameters and filepaths.
type PrivValidatorConfig struct {
	// File path configuration.
	RootDir     string `json:"home" toml:"home"`
	SignState   string `json:"sign_state" toml:"sign_state" comment:"Path to the JSON file containing the last validator state to prevent double-signing"`
	LocalSigner string `json:"local_signer" toml:"local_signer" comment:"Path to the JSON file containing the private key to use for signing using a local signer"`

	// Remote Signer configuration.
	RemoteSigner *rsclient.RemoteSignerClientConfig `json:"remote_signer" toml:"remote_signer" comment:"Configuration for the remote signer client"`
}

// PrivValidatorConfig validation errors.
var (
	errInvalidSignStatePath   = errors.New("invalid private validator sign state file path")
	errInvalidLocalSignerPath = errors.New("invalid private validator local signer file path")
)

// DefaultPrivValidatorConfig returns a default configuration for the PrivValidator.
func DefaultPrivValidatorConfig() *PrivValidatorConfig {
	return &PrivValidatorConfig{
		SignState:    "priv_validator_state.json",
		LocalSigner:  "priv_validator_key.json",
		RemoteSigner: rsclient.DefaultRemoteSignerClientConfig(),
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

	// Validate the remote signer client configuration.
	if err := cfg.RemoteSigner.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// NewPrivValidatorFromConfig returns a new PrivValidator instance based on the configuration.
// The clientLogger is only used for the remote signer client and ignored it the signer is local.
// The clientPrivKey is only used for the remote signer client using a TCP connection.
func NewPrivValidatorFromConfig(
	config *PrivValidatorConfig,
	clientPrivKey ed25519.PrivKeyEd25519,
	clientLogger *slog.Logger,
) (*PrivValidator, error) {
	var (
		signer types.Signer
		err    error
	)

	// Initialize the signer based on the configuration.
	// If the remote signer address is set, use a remote signer client.
	if config.RemoteSigner != nil && config.RemoteSigner.ServerAddress != "" {
		signer, err = rsclient.NewRemoteSignerClientFromConfig(
			config.RemoteSigner,
			clientPrivKey,
			clientLogger,
		)
	} else {
		// Otherwise, use a local signer.
		signer, err = local.LoadOrMakeLocalSigner(config.LocalSignerPath())
	}
	if err != nil {
		return nil, fmt.Errorf("signer initialization from config failed: %w", err)
	}

	return NewPrivValidator(signer, config.SignStatePath())
}
