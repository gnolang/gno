package vault

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockKV is an in-memory fake implementing kvAPI, used so tests never make
// real Vault API calls or require a running Vault server.
type mockKV struct {
	// secrets maps secret path -> stored data map.
	secrets map[string]map[string]interface{}

	getErr error
	putErr error
}

func newMockKV() *mockKV {
	return &mockKV{secrets: map[string]map[string]interface{}{}}
}

func (m *mockKV) Get(_ context.Context, secretPath string) (*vaultapi.KVSecret, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}

	data, ok := m.secrets[secretPath]
	if !ok {
		return nil, vaultapi.ErrSecretNotFound
	}

	return &vaultapi.KVSecret{Data: data}, nil
}

func (m *mockKV) Put(
	_ context.Context,
	secretPath string,
	data map[string]interface{},
	_ ...vaultapi.KVOption,
) (*vaultapi.KVSecret, error) {
	if m.putErr != nil {
		return nil, m.putErr
	}

	m.secrets[secretPath] = data

	return &vaultapi.KVSecret{Data: data}, nil
}

func TestNewSigner_ExistingSecret(t *testing.T) {
	t.Parallel()

	key := local.GenerateFileKey()
	jsonBytes, err := amino.MarshalJSONIndent(key, "", "  ")
	require.NoError(t, err)

	client := newMockKV()
	client.secrets["gno/validator-key"] = map[string]interface{}{
		dataFieldName: string(jsonBytes),
	}

	cfg := &Config{SecretPath: "gno/validator-key"}

	signer, err := newSigner(context.Background(), client, cfg)
	require.NoError(t, err)
	require.NotNil(t, signer)

	assert.True(t, signer.PubKey().Equals(key.PubKey))

	sig, err := signer.Sign([]byte("hello"))
	require.NoError(t, err)
	assert.True(t, signer.PubKey().VerifyBytes([]byte("hello"), sig))

	assert.NoError(t, signer.Close())
	assert.Contains(t, signer.String(), "VaultSigner")
}

func TestNewSigner_MissingSecret_NoCreate(t *testing.T) {
	t.Parallel()

	client := newMockKV()
	cfg := &Config{SecretPath: "does/not/exist", CreateIfMissing: false}

	signer, err := newSigner(context.Background(), client, cfg)
	require.Error(t, err)
	assert.Nil(t, signer)
}

func TestNewSigner_MissingSecret_CreateIfMissing(t *testing.T) {
	t.Parallel()

	client := newMockKV()
	cfg := &Config{SecretPath: "gno/new-validator-key", CreateIfMissing: true}

	signer, err := newSigner(context.Background(), client, cfg)
	require.NoError(t, err)
	require.NotNil(t, signer)

	stored, ok := client.secrets["gno/new-validator-key"]
	require.True(t, ok)

	rawStr, ok := stored[dataFieldName].(string)
	require.True(t, ok)

	parsed, err := local.ParseFileKey([]byte(rawStr))
	require.NoError(t, err)
	assert.True(t, signer.PubKey().Equals(parsed.PubKey))
}

func TestNewSigner_MalformedSecret(t *testing.T) {
	t.Parallel()

	client := newMockKV()
	client.secrets["gno/bad-key"] = map[string]interface{}{
		dataFieldName: "{not valid json",
	}

	cfg := &Config{SecretPath: "gno/bad-key"}

	signer, err := newSigner(context.Background(), client, cfg)
	require.Error(t, err)
	assert.Nil(t, signer)
}

func TestNewSigner_MissingDataField(t *testing.T) {
	t.Parallel()

	client := newMockKV()
	client.secrets["gno/wrong-field"] = map[string]interface{}{
		"some_other_field": "value",
	}

	cfg := &Config{SecretPath: "gno/wrong-field"}

	signer, err := newSigner(context.Background(), client, cfg)
	require.ErrorIs(t, err, errMissingDataField)
	assert.Nil(t, signer)
}

func TestConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	assert.False(t, (&Config{}).IsEnabled())
	assert.False(t, (*Config)(nil).IsEnabled())
	assert.True(t, (&Config{SecretPath: "x"}).IsEnabled())
}

func TestConfig_MountPath(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "secret", (&Config{}).mountPath())
	assert.Equal(t, "custom", (&Config{MountPath: "custom"}).mountPath())
}

func TestNewSignerFromConfig_Disabled(t *testing.T) {
	t.Parallel()

	_, err := NewSignerFromConfig(context.Background(), DefaultConfig())
	require.ErrorIs(t, err, errSecretPathRequired)
}
