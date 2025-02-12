package local

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLocalSigner(t *testing.T) {
	t.Parallel()

	t.Run("load existing signer's key", func(t *testing.T) {
		t.Parallel()

		// Generate a valid random local signer (FileKey) on disk.
		filePath := path.Join(t.TempDir(), "existing")
		ls, err := NewLocalSigner(filePath)
		require.NoError(t, err)

		// Load it using NewLocalSigner.
		loaded, err := NewLocalSigner(filePath)
		require.NotNil(t, loaded)
		require.NoError(t, err)

		// Compare the loaded file key with the original.
		require.Equal(t, ls.key, loaded.key)
	})

	t.Run("read-only file path", func(t *testing.T) {
		t.Parallel()

		// Create a read-only directory.
		dirPath := path.Join(t.TempDir(), "read-only")
		err := os.Mkdir(dirPath, 0444)
		require.NoError(t, err)

		filePath := path.Join(dirPath, "file")
		fk, err := NewLocalSigner(filePath)
		require.Nil(t, fk)
		require.Error(t, err)
	})

	t.Run("simple valid flow", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "new")
		fk, err := NewLocalSigner(filePath)
		require.NotNil(t, fk)
		require.NoError(t, err)

		signBytes := []byte("sign bytes")
		signature, err := fk.Sign(signBytes)
		require.NotNil(t, signature)
		require.NoError(t, err)

		pk, err := fk.PubKey()
		require.NotNil(t, pk)
		require.NoError(t, err)

		require.True(t, pk.VerifyBytes(signBytes, signature))
		require.Contains(t, fk.String(), fk.key.Address.String())
	})
}
