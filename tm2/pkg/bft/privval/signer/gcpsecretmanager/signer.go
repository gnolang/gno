package gcpsecretmanager

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Signer implements types.Signer using a validator key stored in GCP Secret
// Manager. The key material is fetched once at construction time and kept in
// memory for the lifetime of the process; signing itself happens locally,
// the same way the local file signer does. GCP Secret Manager is used purely
// as a durable, access-controlled key store, not as a remote signing oracle —
// unlike the tmkms/gnokms remote-signer modes, the private key does leave
// GCP and reside in this process's memory.
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
	return fmt.Sprintf("{Type: GCPSecretManagerSigner, Addr: %s}", s.key.Address)
}

// Config validation errors.
var (
	errConfigDisabled   = errors.New("gcp secret manager signer: project_id and secret_id are required")
	errEmptySecretValue = errors.New("gcp secret manager signer: secret has no value")
)

// NewSignerFromConfig fetches the validator key from GCP Secret Manager
// according to cfg and returns a ready-to-use Signer. If the secret does not
// exist and cfg.CreateIfMissing is true, a new random key is generated and
// stored under cfg.SecretID (with an initial version) before being returned.
func NewSignerFromConfig(ctx context.Context, cfg *Config) (*Signer, error) {
	if !cfg.IsEnabled() {
		return nil, errConfigDisabled
	}

	client, err := newClient(ctx)
	if err != nil {
		return nil, err
	}

	return newSigner(ctx, client, cfg)
}

// newSigner contains the constructor logic decoupled from the concrete GCP
// client, so it can be exercised in tests against a mock secretManagerAPI.
func newSigner(ctx context.Context, client secretManagerAPI, cfg *Config) (*Signer, error) {
	out, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: cfg.versionName(),
	})
	if err != nil {
		if cfg.CreateIfMissing && isNotFound(err) {
			return createAndStoreKey(ctx, client, cfg)
		}

		return nil, fmt.Errorf("unable to access secret %q from GCP Secret Manager: %w", cfg.versionName(), err)
	}

	if out.Payload == nil || len(out.Payload.Data) == 0 {
		return nil, errEmptySecretValue
	}

	key, err := local.ParseFileKey(out.Payload.Data)
	if err != nil {
		return nil, fmt.Errorf("invalid validator key in secret %q: %w", cfg.versionName(), err)
	}

	return &Signer{key: key}, nil
}

// isNotFound reports whether err represents a gRPC NotFound status, which
// GCP Secret Manager returns both when the secret itself doesn't exist and
// when it exists but has no versions yet.
func isNotFound(err error) bool {
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.NotFound
}

// createAndStoreKey generates a new random validator FileKey, creates a new
// secret for it (with automatic replication) and adds an initial version
// containing the same amino-JSON encoding as the local file signer, then
// returns a Signer wrapping the generated key.
func createAndStoreKey(ctx context.Context, client secretManagerAPI, cfg *Config) (*Signer, error) {
	key := local.GenerateFileKey()

	jsonBytes, err := amino.MarshalJSONIndent(key, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("unable to marshal newly generated validator key: %w", err)
	}

	secret, err := client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
		Parent:   cfg.secretParent(),
		SecretId: cfg.SecretID,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create secret %q in GCP Secret Manager: %w", cfg.SecretID, err)
	}

	if _, err := client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: jsonBytes,
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to add secret version for %q in GCP Secret Manager: %w", cfg.SecretID, err)
	}

	return &Signer{key: key}, nil
}
