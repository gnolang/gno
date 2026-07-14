package vault

// Config defines the configuration options for a Signer backed by
// HashiCorp Vault's KV v2 secrets engine. This is used to marshal/unmarshal
// the configuration to/from TOML and configure the signer using the gnoland
// CLI tool.
type Config struct {
	// Address is the Vault server address (e.g. "https://vault.example.com:8200").
	// If empty, the client falls back to the standard Vault SDK resolution
	// chain (the VAULT_ADDR environment variable, then http://127.0.0.1:8200).
	Address string `json:"address" toml:"address" comment:"Vault server address. Leave empty to use the VAULT_ADDR environment variable / SDK default"`

	// Token is the Vault token used to authenticate. If empty, the client
	// falls back to the standard Vault SDK resolution chain (the VAULT_TOKEN
	// environment variable, or the ~/.vault-token file written by `vault login`).
	Token string `json:"token" toml:"token" comment:"Vault token. Leave empty to use the VAULT_TOKEN environment variable / ~/.vault-token"`

	// MountPath is the mount path of the KV v2 secrets engine holding the
	// secret. Defaults to "secret" (the standard KV v2 mount) if empty.
	MountPath string `json:"mount_path" toml:"mount_path" comment:"Mount path of the KV v2 secrets engine. Defaults to \"secret\" if empty"`

	// SecretPath is the path (within MountPath) of the secret holding the
	// validator's private key, encoded the same way as the on-disk
	// priv_validator_key.json file. If empty, the Vault signer is disabled.
	SecretPath string `json:"secret_path" toml:"secret_path" comment:"Path (within mount_path) of the secret holding the validator key. If set, the local signer is disabled"`

	// CreateIfMissing, when true, generates a new random validator key and
	// writes it to SecretPath if no secret exists there yet. When false
	// (the default), a missing secret is treated as a fatal configuration
	// error, to avoid silently minting a validator identity that was never
	// intended.
	CreateIfMissing bool `json:"create_if_missing" toml:"create_if_missing" comment:"Generate and store a new validator key in Vault if the secret does not exist yet"`
}

// DefaultConfig returns a default, disabled configuration for the Vault signer.
func DefaultConfig() *Config {
	return &Config{
		Address:         "",
		Token:           "",
		MountPath:       "secret",
		SecretPath:      "", // Empty to disable the Vault signer by default.
		CreateIfMissing: false,
	}
}

// TestConfig returns a configuration for testing the Vault signer.
func TestConfig() *Config {
	return DefaultConfig()
}

// IsEnabled reports whether the Vault signer is configured for use.
func (cfg *Config) IsEnabled() bool {
	return cfg != nil && cfg.SecretPath != ""
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *Config) ValidateBasic() error {
	// No cross-field constraints beyond SecretPath gating IsEnabled: Address,
	// Token, MountPath, and CreateIfMissing are all meaningful at their zero
	// value (falling back to the Vault SDK's own resolution chain).
	return nil
}

// mountPath returns the configured mount path, defaulting to "secret" (the
// standard KV v2 mount) if unset.
func (cfg *Config) mountPath() string {
	if cfg.MountPath == "" {
		return "secret"
	}

	return cfg.MountPath
}
