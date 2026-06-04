package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// RemoteSignerClientConfig defines the configuration options for a RemoteSignerClient.
// This is used to marshal/unmarshal the configuration to/from TOML and configure the client
// using the gnoland CLI tool.
type RemoteSignerClientConfig struct {
	// Address of the remote signer to dial (UNIX or TCP).
	ServerAddress string `json:"server_address" toml:"server_address" comment:"Address of the remote signer to dial (UNIX or TCP). If set, the local signer is disabled"`

	// Network dial and timeout options.
	DialMaxRetries    int           `json:"dial_max_retries" toml:"dial_max_retries" comment:"Maximum number of retries to dial the remote signer. If set to -1, will retry indefinitely"`
	DialRetryInterval time.Duration `json:"dial_retry_interval" toml:"dial_retry_interval" comment:"Interval between retries to dial the remote signer"`
	DialTimeout       time.Duration `json:"dial_timeout" toml:"dial_timeout" comment:"Timeout to dial the remote signer"`
	RequestTimeout    time.Duration `json:"request_timeout" toml:"request_timeout" comment:"Timeout for requests to the remote signer"`

	// TCP specific options.
	AuthorizedKeys  []string      `json:"tcp_authorized_keys" toml:"tcp_authorized_keys" comment:"List of authorized public keys for the remote signer (only for TCP). If empty, all keys are authorized"`
	KeepAlivePeriod time.Duration `json:"tcp_keep_alive_period" toml:"tcp_keep_alive_period" comment:"Keep alive period for the remote signer connection (only for TCP)"`
}

// DefaultRemoteSignerClientConfig returns a default configuration for the RemoteSignerClient.
func DefaultRemoteSignerClientConfig() *RemoteSignerClientConfig {
	return &RemoteSignerClientConfig{
		ServerAddress:     "", // Empty to disable remote signer by default.
		DialMaxRetries:    defaultDialMaxRetries,
		DialRetryInterval: defaultDialRetryInterval,
		DialTimeout:       defaultDialTimeout,
		RequestTimeout:    defaultRequestTimeout,
		KeepAlivePeriod:   defaultKeepAlivePeriod,
		AuthorizedKeys:    []string{}, // Empty to authorize all keys by default.
	}
}

// TestRemoteSignerClientConfig returns a configuration for testing the RemoteSignerClient.
func TestRemoteSignerClientConfig() *RemoteSignerClientConfig {
	return DefaultRemoteSignerClientConfig()
}

// RemoteSignerClientConfig validation errors.
var errInvalidAuthorizedKey = errors.New("invalid remote signer authorized key")

// authorizedKeys returns the authorized public keys for the remote signer in ed25519
// format and returns an error if any key is invalid.
func (cfg *RemoteSignerClientConfig) authorizedKeys() ([]ed25519.PubKeyEd25519, error) {
	keys := make([]ed25519.PubKeyEd25519, len(cfg.AuthorizedKeys))

	for i := range cfg.AuthorizedKeys {
		// Decode the public key from the Bech32 format.
		pubKey, err := crypto.PubKeyFromBech32(cfg.AuthorizedKeys[i])
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
func (cfg *RemoteSignerClientConfig) ValidateBasic() error {
	// Verify the authorized keys are valid.
	if _, err := cfg.authorizedKeys(); err != nil {
		return err
	}

	return nil
}

// NewRemoteSignerClientFromConfig returns a new RemoteSignerClient instance based on the configuration.
// The clientPrivKey is only used if the client connects to the server using TCP.
func NewRemoteSignerClientFromConfig(
	ctx context.Context,
	config *RemoteSignerClientConfig,
	clientPrivKey ed25519.PrivKeyEd25519,
	clientLogger *slog.Logger,
) (*RemoteSignerClient, error) {
	// Options for the remote signer client.
	options := []Option{
		WithClientPrivKey(clientPrivKey),
		WithDialMaxRetries(config.DialMaxRetries),
		WithDialRetryInterval(config.DialRetryInterval),
		WithDialTimeout(config.DialTimeout),
		WithRequestTimeout(config.RequestTimeout),
	}

	// If authorized keys are set in the config, add them to the options.
	if len(config.AuthorizedKeys) > 0 {
		authorizedKeys, err := config.authorizedKeys()
		if err != nil {
			return nil, err
		}
		options = append(options, WithAuthorizedKeys(authorizedKeys))
	}

	return NewRemoteSignerClient(ctx, config.ServerAddress, clientLogger, options...)
}
