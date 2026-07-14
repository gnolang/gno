package gcpsecretmanager

import "fmt"

// Config defines the configuration options for a Signer backed by GCP Secret
// Manager. This is used to marshal/unmarshal the configuration to/from TOML
// and configure the signer using the gnoland CLI tool.
type Config struct {
	// ProjectID is the GCP project that owns the secret. If empty, the GCP
	// Secret Manager signer is disabled.
	ProjectID string `json:"project_id" toml:"project_id" comment:"GCP project ID that owns the secret"`

	// SecretID is the short ID (not the full resource name) of the GCP
	// Secret Manager secret holding the validator's private key, encoded
	// the same way as the on-disk priv_validator_key.json file. If empty,
	// the GCP Secret Manager signer is disabled.
	SecretID string `json:"secret_id" toml:"secret_id" comment:"Short ID of the GCP Secret Manager secret holding the validator key. If set (together with project_id), the local signer is disabled"`

	// Version is the secret version to read (e.g. "1", or "latest"). If
	// empty, "latest" is used.
	Version string `json:"version" toml:"version" comment:"Secret version to read. Defaults to \"latest\" if empty"`

	// CreateIfMissing, when true, generates a new random validator key and
	// stores it as a new secret (with an initial version) under SecretID if
	// the secret does not already exist. When false (the default), a
	// missing secret is treated as a fatal configuration error, to avoid
	// silently minting a validator identity that was never intended.
	CreateIfMissing bool `json:"create_if_missing" toml:"create_if_missing" comment:"Generate and store a new validator key in Secret Manager if the secret does not exist yet"`
}

// DefaultConfig returns a default, disabled configuration for the GCP Secret
// Manager signer.
func DefaultConfig() *Config {
	return &Config{
		ProjectID:       "", // Empty to disable the GCP Secret Manager signer by default.
		SecretID:        "",
		Version:         "latest",
		CreateIfMissing: false,
	}
}

// TestConfig returns a configuration for testing the GCP Secret Manager signer.
func TestConfig() *Config {
	return DefaultConfig()
}

// IsEnabled reports whether the GCP Secret Manager signer is configured for use.
func (cfg *Config) IsEnabled() bool {
	return cfg != nil && cfg.ProjectID != "" && cfg.SecretID != ""
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *Config) ValidateBasic() error {
	// No cross-field constraints beyond ProjectID+SecretID gating IsEnabled:
	// Version and CreateIfMissing are both meaningful at their zero value.
	return nil
}

// versionName returns the fully-qualified resource name of the secret
// version to access, e.g. "projects/my-project/secrets/my-secret/versions/latest".
func (cfg *Config) versionName() string {
	version := cfg.Version
	if version == "" {
		version = "latest"
	}

	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", cfg.ProjectID, cfg.SecretID, version)
}

// secretParent returns the fully-qualified resource name of the parent
// project, used when creating a new secret.
func (cfg *Config) secretParent() string {
	return fmt.Sprintf("projects/%s", cfg.ProjectID)
}
