package local

import (
	"os"
	"path"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid FileKey", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey()

		assert.NoError(t, fk.validate())
	})

	t.Run("invalid private key", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey()
		fk.PrivKey = nil

		assert.ErrorIs(t, fk.validate(), errInvalidPrivateKey)
	})

	t.Run("public key mismatch", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey()
		fk.PubKey = nil

		assert.ErrorIs(t, fk.validate(), errPublicKeyMismatch)
	})

	t.Run("address mismatch", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey()
		fk.Address = crypto.Address{} // zero address

		assert.ErrorIs(t, fk.validate(), errAddressMismatch)
	})
}

func TestSave(t *testing.T) {
	t.Parallel()

	t.Run("empty file path", func(t *testing.T) {
		t.Parallel()

		fk, err := GeneratePersistedFileKey("")
		require.Nil(t, fk)
		assert.Error(t, err)
	})

	t.Run("read-only file path", func(t *testing.T) {
		t.Parallel()

		// Create a read-only directory.
		dirPath := path.Join(t.TempDir(), "read-only")
		err := os.Mkdir(dirPath, 0o444)
		require.NoError(t, err)

		filePath := path.Join(dirPath, "file")
		fk, err := GeneratePersistedFileKey(filePath)
		require.Nil(t, fk)
		assert.Error(t, err)
	})

	t.Run("read-write file path", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "writable")
		fk, err := GeneratePersistedFileKey(filePath)
		require.NotNil(t, fk)
		assert.NoError(t, err)
	})
}

func TestLoadFileKey(t *testing.T) {
	t.Parallel()

	t.Run("valid file key", func(t *testing.T) {
		t.Parallel()

		// Generate a valid random file key on disk.
		filePath := path.Join(t.TempDir(), "valid")
		fk, err := GeneratePersistedFileKey(filePath)
		require.NoError(t, err)

		// Load the file key from disk.
		loaded, err := LoadFileKey(filePath)
		require.NoError(t, err)

		// Compare the loaded file key with the original.
		assert.Equal(t, fk, loaded)
	})

	t.Run("non-existent file path", func(t *testing.T) {
		t.Parallel()

		fk, err := LoadFileKey("non-existent")
		require.Nil(t, fk)
		assert.Error(t, err)
	})

	t.Run("invalid file key", func(t *testing.T) {
		t.Parallel()

		// Create a file with invalid FileKey JSON.
		filePath := path.Join(t.TempDir(), "invalid")
		os.WriteFile(filePath, []byte(`{address:"invalid"}`), 0o644)

		fk, err := LoadFileKey(filePath)
		require.Nil(t, fk)
		require.Error(t, err)

		// Generate a valid FileKey first.
		fk, err = GeneratePersistedFileKey(filePath)
		require.NotNil(t, fk)
		require.NoError(t, err)

		// Make its address invalid then persist it to disk.
		copy(fk.Address[:], "invalid address")
		jsonBytes, err := amino.MarshalJSONIndent(fk, "", "  ")
		require.NoError(t, err)
		require.NoError(t, osm.WriteFileAtomic(filePath, jsonBytes, 0o600))

		fk, err = LoadFileKey(filePath)
		require.Nil(t, fk)
		assert.ErrorIs(t, err, errAddressMismatch)
	})
}

func TestNewFileKey(t *testing.T) {
	t.Parallel()

	t.Run("genetate new key", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "new")
		fk, err := LoadOrMakeFileKey(filePath)
		require.NotNil(t, fk)
		assert.NoError(t, err)
	})

	t.Run("load existing key", func(t *testing.T) {
		t.Parallel()

		// Generate a valid random file key on disk.
		filePath := path.Join(t.TempDir(), "existing")
		fk, err := GeneratePersistedFileKey(filePath)
		require.NoError(t, err)

		// Load it using NewFileKey.
		loaded, err := LoadOrMakeFileKey(filePath)
		require.NotNil(t, loaded)
		require.NoError(t, err)

		// Compare the loaded file key with the original.
		assert.Equal(t, fk, loaded)
	})

	t.Run("read-only file path", func(t *testing.T) {
		t.Parallel()

		// Create a read-only directory.
		dirPath := path.Join(t.TempDir(), "read-only")
		err := os.Mkdir(dirPath, 0o444)
		require.NoError(t, err)

		filePath := path.Join(dirPath, "file")
		fk, err := LoadOrMakeFileKey(filePath)
		require.Nil(t, fk)
		assert.Error(t, err)
	})
}
