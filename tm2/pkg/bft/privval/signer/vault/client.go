package vault

import (
	"context"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

// kvAPI is the subset of the Vault KV v2 client used by the signer. It is
// defined as an interface so tests can substitute a mock implementation
// without making real Vault API calls or requiring a running Vault server.
type kvAPI interface {
	Get(ctx context.Context, secretPath string) (*vaultapi.KVSecret, error)
	Put(ctx context.Context, secretPath string, data map[string]interface{}, opts ...vaultapi.KVOption) (*vaultapi.KVSecret, error)
}

// kvAPI is implemented by *vaultapi.KVv2.
var _ kvAPI = (*vaultapi.KVv2)(nil)

// newClient builds a real Vault KV v2 client. Address and Token, if set in
// cfg, override the Vault SDK's own environment/file-based resolution chain
// (VAULT_ADDR, VAULT_TOKEN, ~/.vault-token).
func newClient(cfg *Config) (kvAPI, error) {
	vconfig := vaultapi.DefaultConfig()
	if cfg.Address != "" {
		vconfig.Address = cfg.Address
	}

	c, err := vaultapi.NewClient(vconfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create Vault client: %w", err)
	}

	if cfg.Token != "" {
		c.SetToken(cfg.Token)
	}

	return c.KVv2(cfg.mountPath()), nil
}
