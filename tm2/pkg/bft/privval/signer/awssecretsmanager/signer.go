package awssecretsmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// Signer implements types.Signer using a validator key stored in AWS Secrets
// Manager. The key material is fetched once at construction time and kept in
// memory for the lifetime of the process; signing itself happens locally, the
// same way the local file signer does. AWS Secrets Manager is used purely as
// a durable, access-controlled key store, not as a remote signing oracle —
// unlike the tmkms/gnokms remote-signer modes, the private key does leave
// AWS and reside in this process's memory.
type Signer struct {
	key *local.FileKey
}

// Signer type implements types.Signer.
var _ types.Signer = (*Signer)(nil)

// PubKey implements types.Signer.
func (s *Signer) PubKey() crypto.PubKey {
	return s.key.PubKey
}

// Sign implements types.Signer.
func (s *Signer) Sign(signBytes []byte) ([]byte, error) {
	return s.key.PrivKey.Sign(signBytes)
}

// Close implements types.Signer.
func (s *Signer) Close() error {
	return nil
}

// Signer type implements fmt.Stringer.
var _ fmt.Stringer = (*Signer)(nil)

// String implements fmt.Stringer.
func (s *Signer) String() string {
	return fmt.Sprintf("{Type: AWSSecretsManagerSigner, Addr: %s}", s.key.Address)
}

// Config validation errors.
var (
	errSecretIDRequired = errors.New("aws secrets manager signer: secret_id is required")
	errEmptySecretValue = errors.New("aws secrets manager signer: secret has no value")
)

// NewSignerFromConfig fetches the validator key from AWS Secrets Manager
// according to cfg and returns a ready-to-use Signer. If the secret does not
// exist and cfg.CreateIfMissing is true, a new random key is generated and
// stored under cfg.SecretID before being returned.
func NewSignerFromConfig(ctx context.Context, cfg *Config) (*Signer, error) {
	if !cfg.IsEnabled() {
		return nil, errSecretIDRequired
	}

	client, err := newClient(ctx, cfg.Region)
	if err != nil {
		return nil, err
	}

	return newSigner(ctx, client, cfg)
}

// newSigner contains the constructor logic decoupled from the concrete AWS
// client, so it can be exercised in tests against a mock secretsManagerAPI.
func newSigner(ctx context.Context, client secretsManagerAPI, cfg *Config) (*Signer, error) {
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &cfg.SecretID,
	})
	if err != nil {
		var notFound *smtypes.ResourceNotFoundException
		if errors.As(err, &notFound) && cfg.CreateIfMissing {
			return createAndStoreKey(ctx, client, cfg.SecretID)
		}

		return nil, fmt.Errorf("unable to retrieve secret %q from AWS Secrets Manager: %w", cfg.SecretID, err)
	}

	raw, err := secretPayload(out)
	if err != nil {
		return nil, err
	}

	key, err := local.ParseFileKey(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid validator key in secret %q: %w", cfg.SecretID, err)
	}

	return &Signer{key: key}, nil
}

// secretPayload extracts the raw JSON bytes from a GetSecretValueOutput,
// preferring SecretString (the common case, and what CreateSecret in this
// package writes) and falling back to SecretBinary.
func secretPayload(out *secretsmanager.GetSecretValueOutput) ([]byte, error) {
	if out.SecretString != nil {
		return []byte(*out.SecretString), nil
	}

	if len(out.SecretBinary) > 0 {
		return out.SecretBinary, nil
	}

	return nil, errEmptySecretValue
}

// createAndStoreKey generates a new random validator FileKey, persists it as
// a new secret under secretID using the same amino-JSON encoding as the local
// file signer, and returns a Signer wrapping it.
func createAndStoreKey(ctx context.Context, client secretsManagerAPI, secretID string) (*Signer, error) {
	key := local.GenerateFileKey()

	jsonBytes, err := amino.MarshalJSONIndent(key, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("unable to marshal newly generated validator key: %w", err)
	}
	secretString := string(jsonBytes)

	if _, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         &secretID,
		SecretString: &secretString,
	}); err != nil {
		return nil, fmt.Errorf("unable to create secret %q in AWS Secrets Manager: %w", secretID, err)
	}

	return &Signer{key: key}, nil
}
