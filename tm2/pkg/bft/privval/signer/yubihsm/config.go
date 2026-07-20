package yubihsm

import (
	"errors"
	"os"
)

// Config defines the configuration options for a Signer backed by a
// YubiHSM2 hardware security module, accessed via a yubihsm-connector
// service.
type Config struct {
	// ConnectorURL is the address of the yubihsm-connector service that
	// bridges this process to the physical YubiHSM2 device, as a bare
	// "host:port" (e.g. "127.0.0.1:12345") — the client prepends the
	// scheme itself, so a URL with "http://" already included will break.
	// If empty, the YubiHSM2 signer is disabled.
	ConnectorURL string `json:"connector_url" toml:"connector_url" comment:"Address of the yubihsm-connector service as host:port, e.g. 127.0.0.1:12345 (no scheme). If set, the local signer is disabled"`

	// AuthKeyID is the Object ID of the Authentication Key used to open a
	// session with the device.
	AuthKeyID uint16 `json:"auth_key_id" toml:"auth_key_id" comment:"Object ID of the Authentication Key used to open a session with the device"`

	// Password authenticates against AuthKeyID. Storing it directly here
	// means it lands in config.toml in plaintext; prefer PasswordEnv where
	// possible so the secret isn't persisted to disk alongside the rest of
	// the (otherwise non-sensitive) node configuration.
	Password string `json:"password" toml:"password" comment:"Password for the Authentication Key. Prefer password_env over this where possible"`

	// PasswordEnv, if set, names an environment variable to read the
	// Authentication Key password from at startup, taking precedence over
	// Password. This keeps the secret out of config.toml.
	PasswordEnv string `json:"password_env" toml:"password_env" comment:"Name of an environment variable to read the password from (takes precedence over 'password')"`

	// KeyID is the Object ID of the Ed25519 asymmetric key on the device
	// holding the validator's private key.
	KeyID uint16 `json:"key_id" toml:"key_id" comment:"Object ID of the Ed25519 asymmetric key on the device holding the validator key"`
}

// DefaultConfig returns a default, disabled configuration for the YubiHSM2 signer.
func DefaultConfig() *Config {
	return &Config{
		ConnectorURL: "", // Empty to disable the YubiHSM2 signer by default.
		AuthKeyID:    0,
		Password:     "",
		PasswordEnv:  "",
		KeyID:        0,
	}
}

// TestConfig returns a configuration for testing the YubiHSM2 signer.
func TestConfig() *Config {
	return DefaultConfig()
}

// IsEnabled reports whether the YubiHSM2 signer is configured for use.
func (cfg *Config) IsEnabled() bool {
	return cfg != nil && cfg.ConnectorURL != ""
}

// Config validation errors.
var (
	errZeroAuthKeyID       = errors.New("yubihsm signer: auth_key_id cannot be zero when enabled")
	errZeroKeyID           = errors.New("yubihsm signer: key_id cannot be zero when enabled")
	errConnectorURLMissing = errors.New(
		"yubihsm signer: connector_url must be set if auth_key_id, key_id, or a password is configured; " +
			"otherwise the signer silently falls back to the local file signer")
)

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *Config) ValidateBasic() error {
	if cfg == nil {
		return nil
	}

	// Guard against a partially-filled-but-technically-disabled config: if
	// any other yubihsm field is set but connector_url isn't, IsEnabled
	// reports false and the node would silently fall back to the local
	// file signer instead of the YubiHSM2 the operator clearly intended to
	// configure. Fail loudly instead.
	if cfg.ConnectorURL == "" {
		if cfg.AuthKeyID != 0 || cfg.KeyID != 0 || cfg.Password != "" || cfg.PasswordEnv != "" {
			return errConnectorURLMissing
		}
		return nil
	}

	if cfg.AuthKeyID == 0 {
		return errZeroAuthKeyID
	}

	if cfg.KeyID == 0 {
		return errZeroKeyID
	}

	return nil
}

// resolvePassword returns the Authentication Key password to use: the
// value of the PasswordEnv environment variable if PasswordEnv is set,
// otherwise the literal Password field.
func (cfg *Config) resolvePassword() string {
	if cfg.PasswordEnv != "" {
		return os.Getenv(cfg.PasswordEnv)
	}

	return cfg.Password
}
