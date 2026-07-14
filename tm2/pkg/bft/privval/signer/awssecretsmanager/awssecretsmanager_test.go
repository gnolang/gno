package awssecretsmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSecretsManagerClient is an in-memory fake implementing secretsManagerAPI,
// used so tests never make real AWS API calls.
type mockSecretsManagerClient struct {
	// secrets maps secret ID -> stored SecretString value.
	secrets map[string]string

	// getErr, if set, is returned by every GetSecretValue call instead of
	// looking up secrets.
	getErr error

	// createErr, if set, is returned by every CreateSecret call.
	createErr error
}

func newMockClient() *mockSecretsManagerClient {
	return &mockSecretsManagerClient{secrets: map[string]string{}}
}

func (m *mockSecretsManagerClient) GetSecretValue(
	_ context.Context,
	params *secretsmanager.GetSecretValueInput,
	_ ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}

	value, ok := m.secrets[*params.SecretId]
	if !ok {
		msg := "secret not found"
		return nil, &smtypes.ResourceNotFoundException{Message: &msg}
	}

	return &secretsmanager.GetSecretValueOutput{SecretString: &value}, nil
}

func (m *mockSecretsManagerClient) CreateSecret(
	_ context.Context,
	params *secretsmanager.CreateSecretInput,
	_ ...func(*secretsmanager.Options),
) (*secretsmanager.CreateSecretOutput, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}

	m.secrets[*params.Name] = *params.SecretString

	return &secretsmanager.CreateSecretOutput{Name: params.Name}, nil
}

func TestNewSigner_ExistingSecret(t *testing.T) {
	t.Parallel()

	// Seed the mock store with a validator key encoded the same way the
	// local file signer would persist it to disk.
	key := local.GenerateFileKey()
	jsonBytes, err := amino.MarshalJSONIndent(key, "", "  ")
	require.NoError(t, err)

	client := newMockClient()
	client.secrets["my-validator-key"] = string(jsonBytes)

	cfg := &Config{SecretID: "my-validator-key"}

	signer, err := newSigner(context.Background(), client, cfg)
	require.NoError(t, err)
	require.NotNil(t, signer)

	assert.True(t, signer.PubKey().Equals(key.PubKey))

	sig, err := signer.Sign([]byte("hello"))
	require.NoError(t, err)
	assert.True(t, signer.PubKey().VerifyBytes([]byte("hello"), sig))

	assert.NoError(t, signer.Close())
	assert.Contains(t, signer.String(), "AWSSecretsManagerSigner")
}

func TestNewSigner_MissingSecret_NoCreate(t *testing.T) {
	t.Parallel()

	client := newMockClient()
	cfg := &Config{SecretID: "does-not-exist", CreateIfMissing: false}

	signer, err := newSigner(context.Background(), client, cfg)
	require.Error(t, err)
	assert.Nil(t, signer)
}

func TestNewSigner_MissingSecret_CreateIfMissing(t *testing.T) {
	t.Parallel()

	client := newMockClient()
	cfg := &Config{SecretID: "new-validator-key", CreateIfMissing: true}

	signer, err := newSigner(context.Background(), client, cfg)
	require.NoError(t, err)
	require.NotNil(t, signer)

	// The generated key must have been persisted back to the mock store.
	stored, ok := client.secrets["new-validator-key"]
	require.True(t, ok)

	parsed, err := local.ParseFileKey([]byte(stored))
	require.NoError(t, err)
	assert.True(t, signer.PubKey().Equals(parsed.PubKey))
}

func TestNewSigner_MalformedSecret(t *testing.T) {
	t.Parallel()

	client := newMockClient()
	client.secrets["bad-key"] = "{not valid json"

	cfg := &Config{SecretID: "bad-key"}

	signer, err := newSigner(context.Background(), client, cfg)
	require.Error(t, err)
	assert.Nil(t, signer)
}

func TestConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	assert.False(t, (&Config{}).IsEnabled())
	assert.False(t, (*Config)(nil).IsEnabled())
	assert.True(t, (&Config{SecretID: "x"}).IsEnabled())
}

func TestNewSignerFromConfig_Disabled(t *testing.T) {
	t.Parallel()

	_, err := NewSignerFromConfig(context.Background(), DefaultConfig())
	require.ErrorIs(t, err, errSecretIDRequired)
}
