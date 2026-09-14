package local

import (
	"path"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKeyType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    KeyType
		wantErr bool
	}{
		{name: "ed25519", input: "ed25519", want: KeyTypeEd25519},
		{name: "secp256k1", input: "secp256k1", want: KeyTypeSecp256k1},
		{name: "empty", input: "", wantErr: true},
		{name: "unknown", input: "p256", wantErr: true},
		{name: "uppercase rejected", input: "Ed25519", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseKeyType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateFileKeyOfType(t *testing.T) {
	t.Parallel()

	t.Run("ed25519", func(t *testing.T) {
		t.Parallel()

		fk, err := GenerateFileKeyOfType(KeyTypeEd25519)
		require.NoError(t, err)
		require.NotNil(t, fk)

		_, ok := fk.PrivKey.(ed25519.PrivKeyEd25519)
		assert.True(t, ok, "private key should be ed25519")
		assert.NoError(t, fk.validate())
	})

	t.Run("secp256k1", func(t *testing.T) {
		t.Parallel()

		fk, err := GenerateFileKeyOfType(KeyTypeSecp256k1)
		require.NoError(t, err)
		require.NotNil(t, fk)

		_, ok := fk.PrivKey.(secp256k1.PrivKeySecp256k1)
		assert.True(t, ok, "private key should be secp256k1")
		assert.NoError(t, fk.validate())
	})

	t.Run("unsupported", func(t *testing.T) {
		t.Parallel()

		fk, err := GenerateFileKeyOfType(KeyType("p256"))
		assert.Error(t, err)
		assert.Nil(t, fk)
	})
}

func TestGenerateFileKey_DefaultStaysEd25519(t *testing.T) {
	t.Parallel()

	// Backwards-compat guarantee: existing callers using GenerateFileKey
	// without specifying a scheme must continue to get ed25519.
	fk := GenerateFileKey()
	_, ok := fk.PrivKey.(ed25519.PrivKeyEd25519)
	assert.True(t, ok, "default scheme regression: GenerateFileKey returned %T", fk.PrivKey)
}

func TestPersistedFileKey_Secp256k1_RoundTrip(t *testing.T) {
	t.Parallel()

	filePath := path.Join(t.TempDir(), "priv_validator_key.json")

	original, err := GeneratePersistedFileKeyOfType(filePath, KeyTypeSecp256k1)
	require.NoError(t, err)
	require.NotNil(t, original)

	loaded, err := LoadFileKey(filePath)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.True(t, original.PrivKey.Equals(loaded.PrivKey))
	assert.True(t, original.PubKey.Equals(loaded.PubKey))
	assert.Equal(t, original.Address, loaded.Address)

	_, ok := loaded.PrivKey.(secp256k1.PrivKeySecp256k1)
	assert.True(t, ok)
}
