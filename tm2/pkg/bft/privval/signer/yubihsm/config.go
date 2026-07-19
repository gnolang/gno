package yubihsm

import "errors"

// Config defines the configuration options for a Signer backed by a
// YubiHSM2 hardware security module, accessed via a yubihsm-connector
// service.
type Config struct {
	// ConnectorURL is the address of the yubihsm-connector service that
	// bridges this process to the physical YubiHSM2 device (e.g.
	// "http://127.0.0.1:12345"). If empty, the YubiHSM2 signer is disabled.
	ConnectorURL string `json:"connector_url" toml:"connector_url" comment:"Address of the yubihsm-connector service (e.g. http://127.0.0.1:12345). If set, the local signer is disabled"`

	// AuthKeyID is the Object ID of the Authentication Key used to open a
	// session with the device.
	AuthKeyID uint16 `json:"auth_key_id" toml:"auth_key_id" comment:"Object ID of the Authentication Key used to open a session with the device"`

	// Password authenticates against AuthKeyID.
	Password string `json:"password" toml:"password" comment:"Password for the Authentication Key"`

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
	errZeroAuthKeyID = errors.New("yubihsm signer: auth_key_id cannot be zero when enabled")
	errZeroKeyID     = errors.New("yubihsm signer: key_id cannot be zero when enabled")
)

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *Config) ValidateBasic() error {
	if !cfg.IsEnabled() {
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
