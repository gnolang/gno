package awssecretsmanager

// Config defines the configuration options for a Signer backed by AWS
// Secrets Manager. This is used to marshal/unmarshal the configuration
// to/from TOML and configure the signer using the gnoland CLI tool.
type Config struct {
	// SecretID is the ARN or name of the AWS Secrets Manager secret holding
	// the validator's private key, encoded the same way as the on-disk
	// priv_validator_key.json file. If empty, the AWS Secrets Manager signer
	// is disabled.
	SecretID string `json:"secret_id" toml:"secret_id" comment:"ARN or name of the AWS Secrets Manager secret holding the validator key. If set, the local signer is disabled"`

	// Region is the AWS region to use for the Secrets Manager client. If
	// empty, it is resolved using the standard AWS SDK configuration chain
	// (environment variables, shared config/credentials files, EC2/ECS
	// instance metadata, etc.).
	Region string `json:"region" toml:"region" comment:"AWS region for the Secrets Manager client. Leave empty to use the default AWS SDK region resolution"`

	// CreateIfMissing, when true, generates a new random validator key and
	// stores it as a new secret under SecretID if the secret does not
	// already exist. When false (the default), a missing secret is treated
	// as a fatal configuration error, to avoid silently minting a validator
	// identity that was never intended.
	CreateIfMissing bool `json:"create_if_missing" toml:"create_if_missing" comment:"Generate and store a new validator key in Secrets Manager if the secret does not exist yet"`
}

// DefaultConfig returns a default, disabled configuration for the AWS
// Secrets Manager signer.
func DefaultConfig() *Config {
	return &Config{
		SecretID:        "", // Empty to disable the AWS Secrets Manager signer by default.
		Region:          "",
		CreateIfMissing: false,
	}
}

// TestConfig returns a configuration for testing the AWS Secrets Manager signer.
func TestConfig() *Config {
	return DefaultConfig()
}

// IsEnabled reports whether the AWS Secrets Manager signer is configured for use.
func (cfg *Config) IsEnabled() bool {
	return cfg != nil && cfg.SecretID != ""
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *Config) ValidateBasic() error {
	// No cross-field constraints beyond SecretID gating IsEnabled: Region and
	// CreateIfMissing are both meaningful at their zero value.
	return nil
}
