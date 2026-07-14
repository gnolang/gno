package vault

import (
	"context"
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	vaultapi "github.com/hashicorp/vault/api"
)

// dataFieldName is the key under which the validator's FileKey JSON blob is
// stored in the Vault KV v2 secret's data map.
const dataFieldName = "priv_validator_key_json"

// Signer implements types.Signer using a validator key stored in HashiCorp
// Vault's KV v2 secrets engine. The key material is fetched once at
// construction time and kept in memory for the lifetime of the process;
// signing itself happens locally, the same way the local file signer does.
// Vault is used purely as a durable, access-controlled key store, not as a
// remote signing oracle — unlike the tmkms/gnokms remote-signer modes, the
// private key does leave Vault and reside in this process's memory.
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
	return fmt.Sprintf("{Type: VaultSigner, Addr: %s}", s.key.Address)
}

// Config validation errors.
var (
	errSecretPathRequired = errors.New("vault signer: secret_path is required")
	errMissingDataField   = errors.New("vault signer: secret is missing the " + dataFieldName + " field")
	errInvalidDataField   = errors.New("vault signer: " + dataFieldName + " field is not a string")
)

// NewSignerFromConfig fetches the validator key from Vault according to cfg
// and returns a ready-to-use Signer. If the secret does not exist and
// cfg.CreateIfMissing is true, a new random key is generated and written to
// cfg.SecretPath before being returned.
func NewSignerFromConfig(ctx context.Context, cfg *Config) (*Signer, error) {
	if !cfg.IsEnabled() {
		return nil, errSecretPathRequired
	}

	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}

	return newSigner(ctx, client, cfg)
}

// newSigner contains the constructor logic decoupled from the concrete
// Vault client, so it can be exercised in tests against a mock kvAPI.
func newSigner(ctx context.Context, client kvAPI, cfg *Config) (*Signer, error) {
	secret, err := client.Get(ctx, cfg.SecretPath)
	if err != nil {
		if cfg.CreateIfMissing && errors.Is(err, vaultapi.ErrSecretNotFound) {
			return createAndStoreKey(ctx, client, cfg)
		}

		return nil, fmt.Errorf("unable to read secret %q from Vault: %w", cfg.SecretPath, err)
	}

	// A deleted (but not destroyed) KV v2 version returns a nil Data map
	// rather than an error; treat it the same as "not found".
	if secret == nil || secret.Data == nil {
		if cfg.CreateIfMissing {
			return createAndStoreKey(ctx, client, cfg)
		}

		return nil, fmt.Errorf("unable to read secret %q from Vault: %w", cfg.SecretPath, vaultapi.ErrSecretNotFound)
	}

	raw, ok := secret.Data[dataFieldName]
	if !ok {
		return nil, errMissingDataField
	}

	rawStr, ok := raw.(string)
	if !ok {
		return nil, errInvalidDataField
	}

	key, err := local.ParseFileKey([]byte(rawStr))
	if err != nil {
		return nil, fmt.Errorf("invalid validator key in secret %q: %w", cfg.SecretPath, err)
	}

	return &Signer{key: key}, nil
}

// createAndStoreKey generates a new random validator FileKey, writes it to
// Vault under cfg.SecretPath (using the same amino-JSON encoding as the
// local file signer), and returns a Signer wrapping the generated key.
func createAndStoreKey(ctx context.Context, client kvAPI, cfg *Config) (*Signer, error) {
	key := local.GenerateFileKey()

	jsonBytes, err := amino.MarshalJSONIndent(key, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("unable to marshal newly generated validator key: %w", err)
	}

	data := map[string]interface{}{
		dataFieldName: string(jsonBytes),
	}

	if _, err := client.Put(ctx, cfg.SecretPath, data); err != nil {
		return nil, fmt.Errorf("unable to write secret %q to Vault: %w", cfg.SecretPath, err)
	}

	return &Signer{key: key}, nil
}
