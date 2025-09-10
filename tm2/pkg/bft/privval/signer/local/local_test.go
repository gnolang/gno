package local

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringer(t *testing.T) {
	t.Parallel()

	filePath := path.Join(t.TempDir(), "new")
	ls, err := LoadOrMakeLocalSigner(filePath)
	require.NotNil(t, ls)
	require.NoError(t, err)
	assert.Contains(t, ls.String(), ls.key.Address.String())
}

func TestNewLocalSigner(t *testing.T) {
	t.Parallel()

	t.Run("load existing signer's key", func(t *testing.T) {
		t.Parallel()

		// Generate a valid random local signer (FileKey) on disk.
		filePath := path.Join(t.TempDir(), "existing")
		ls, err := LoadOrMakeLocalSigner(filePath)
		require.NoError(t, err)

		// Load it using NewLocalSigner.
		loaded, err := LoadOrMakeLocalSigner(filePath)
		require.NotNil(t, loaded)
		require.NoError(t, err)

		// Compare the loaded file key with the original.
		require.Equal(t, ls.key, loaded.key)
		assert.Nil(t, ls.Close())
	})

	t.Run("read-only file path", func(t *testing.T) {
		t.Parallel()

		// Create a read-only directory.
		dirPath := path.Join(t.TempDir(), "read-only")
		err := os.Mkdir(dirPath, 0o444)
		require.NoError(t, err)

		filePath := path.Join(dirPath, "file")
		ls, err := LoadOrMakeLocalSigner(filePath)
		require.Nil(t, ls)
		require.Error(t, err)
		assert.Nil(t, ls.Close())
	})

	t.Run("simple valid flow", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "new")
		ls, err := LoadOrMakeLocalSigner(filePath)
		require.NotNil(t, ls)
		require.NoError(t, err)

		signBytes := []byte("sign bytes")
		signature, err := ls.Sign(signBytes)
		require.NotNil(t, signature)
		require.NoError(t, err)

		pk := ls.PubKey()
		require.NotNil(t, pk)

		require.True(t, pk.VerifyBytes(signBytes, signature))
		assert.Nil(t, ls.Close())
	})
}
