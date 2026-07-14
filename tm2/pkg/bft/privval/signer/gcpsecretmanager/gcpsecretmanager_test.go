package gcpsecretmanager

import (
	"context"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockClient is an in-memory fake implementing secretManagerAPI, used so
// tests never make real GCP API calls.
type mockClient struct {
	// secrets maps a fully-qualified version resource name -> stored bytes.
	secrets map[string][]byte

	getErr    error
	createErr error
	addErr    error
}

func newMockClient() *mockClient {
	return &mockClient{secrets: map[string][]byte{}}
}

func (m *mockClient) AccessSecretVersion(
	_ context.Context,
	req *secretmanagerpb.AccessSecretVersionRequest,
) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}

	data, ok := m.secrets[req.Name]
	if !ok {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	return &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{Data: data},
	}, nil
}

func (m *mockClient) CreateSecret(
	_ context.Context,
	req *secretmanagerpb.CreateSecretRequest,
) (*secretmanagerpb.Secret, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}

	return &secretmanagerpb.Secret{Name: req.Parent + "/secrets/" + req.SecretId}, nil
}

func (m *mockClient) AddSecretVersion(
	_ context.Context,
	req *secretmanagerpb.AddSecretVersionRequest,
) (*secretmanagerpb.SecretVersion, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}

	// Register under both a numbered version and "latest", mirroring how a
	// subsequent AccessSecretVersion(..., "latest") would resolve on GCP.
	m.secrets[req.Parent+"/versions/1"] = req.Payload.Data
	m.secrets[req.Parent+"/versions/latest"] = req.Payload.Data

	return &secretmanagerpb.SecretVersion{Name: req.Parent + "/versions/1"}, nil
}

func TestNewSigner_ExistingSecret(t *testing.T) {
	t.Parallel()

	key := local.GenerateFileKey()
	jsonBytes, err := amino.MarshalJSONIndent(key, "", "  ")
	require.NoError(t, err)

	cfg := &Config{ProjectID: "my-project", SecretID: "my-validator-key", Version: "latest"}

	client := newMockClient()
	client.secrets[cfg.versionName()] = jsonBytes

	signer, err := newSigner(context.Background(), client, cfg)
	require.NoError(t, err)
	require.NotNil(t, signer)

	assert.True(t, signer.PubKey().Equals(key.PubKey))

	sig, err := signer.Sign([]byte("hello"))
	require.NoError(t, err)
	assert.True(t, signer.PubKey().VerifyBytes([]byte("hello"), sig))

	assert.NoError(t, signer.Close())
	assert.Contains(t, signer.String(), "GCPSecretManagerSigner")
}

func TestNewSigner_MissingSecret_NoCreate(t *testing.T) {
	t.Parallel()

	client := newMockClient()
	cfg := &Config{ProjectID: "my-project", SecretID: "does-not-exist", CreateIfMissing: false}

	signer, err := newSigner(context.Background(), client, cfg)
	require.Error(t, err)
	assert.Nil(t, signer)
}

func TestNewSigner_MissingSecret_CreateIfMissing(t *testing.T) {
	t.Parallel()

	client := newMockClient()
	cfg := &Config{ProjectID: "my-project", SecretID: "new-validator-key", Version: "latest", CreateIfMissing: true}

	signer, err := newSigner(context.Background(), client, cfg)
	require.NoError(t, err)
	require.NotNil(t, signer)

	stored, ok := client.secrets[cfg.versionName()]
	require.True(t, ok)

	parsed, err := local.ParseFileKey(stored)
	require.NoError(t, err)
	assert.True(t, signer.PubKey().Equals(parsed.PubKey))
}

func TestNewSigner_MalformedSecret(t *testing.T) {
	t.Parallel()

	cfg := &Config{ProjectID: "my-project", SecretID: "bad-key"}

	client := newMockClient()
	client.secrets[cfg.versionName()] = []byte("{not valid json")

	signer, err := newSigner(context.Background(), client, cfg)
	require.Error(t, err)
	assert.Nil(t, signer)
}

func TestConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	assert.False(t, (&Config{}).IsEnabled())
	assert.False(t, (*Config)(nil).IsEnabled())
	assert.False(t, (&Config{ProjectID: "p"}).IsEnabled())
	assert.False(t, (&Config{SecretID: "s"}).IsEnabled())
	assert.True(t, (&Config{ProjectID: "p", SecretID: "s"}).IsEnabled())
}

func TestNewSignerFromConfig_Disabled(t *testing.T) {
	t.Parallel()

	_, err := NewSignerFromConfig(context.Background(), DefaultConfig())
	require.ErrorIs(t, err, errConfigDisabled)
}
