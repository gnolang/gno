package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("invalid private key", func(t *testing.T) {
		t.Parallel()

		akf := AuthKeysFile{}

		assert.ErrorIs(t, akf.validate(), errInvalidPublicKey)
	})

	t.Run("public key mismatch", func(t *testing.T) {
		t.Parallel()

		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: ed25519.GenPrivKey(),
				PubKey:  "invalid",
			},
		}

		assert.ErrorIs(t, akf.validate(), errPublicKeyMismatch)
	})

	t.Run("invalid authorized key", func(t *testing.T) {
		t.Parallel()

		privKey := ed25519.GenPrivKey()

		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: privKey,
				PubKey:  privKey.PubKey().String(),
			},
			ClientAuthorizedKeys: []string{"invalid"},
		}

		assert.ErrorIs(t, akf.validate(), errInvalidPublicKey)
	})

	t.Run("valid AuthKeysFile", func(t *testing.T) {
		t.Parallel()

		privKey := ed25519.GenPrivKey()

		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: privKey,
				PubKey:  privKey.PubKey().String(),
			},
			ClientAuthorizedKeys: []string{privKey.PubKey().String()},
		}

		assert.NoError(t, akf.validate())
	})
}

func TestSave(t *testing.T) {
	t.Parallel()

	t.Run("invalid auth keys file", func(t *testing.T) {
		t.Parallel()

		akf := AuthKeysFile{}

		assert.Error(t, akf.Save("filePath"))
	})

	t.Run("invalid parent dir", func(t *testing.T) {
		t.Parallel()

		// Create an unwriteable directory.
		basePath := filepath.Join(t.TempDir(), "unwriteable")
		require.NoError(t, os.MkdirAll(basePath, 0o000))
		filePath := filepath.Join(basePath, "parent", "file")

		privKey := ed25519.GenPrivKey()

		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: privKey,
				PubKey:  privKey.PubKey().String(),
			},
			ClientAuthorizedKeys: []string{privKey.PubKey().String()},
		}

		require.Error(t, akf.Save(filePath))

		// Make the directory unreadable.
		require.NoError(t, os.Chmod(basePath, 0o500))
		require.Error(t, akf.Save(filePath))

		// Restore the permissions for cleanup.
		assert.NoError(t, os.Chmod(basePath, 0o700))
	})

	t.Run("valid auth keys file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(t.TempDir(), "file")
		privKey := ed25519.GenPrivKey()

		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: privKey,
				PubKey:  privKey.PubKey().String(),
			},
			ClientAuthorizedKeys: []string{privKey.PubKey().String()},
		}

		assert.NoError(t, akf.Save(filePath))
	})
}

func TestBech32ToEd25519PubKey(t *testing.T) {
	t.Parallel()

	t.Run("invalid bech32", func(t *testing.T) {
		t.Parallel()

		_, err := Bech32ToEd25519PubKey("invalid")
		assert.Error(t, err)
	})

	t.Run("invalid public key type", func(t *testing.T) {
		t.Parallel()

		privKey := secp256k1.GenPrivKey()

		_, err := Bech32ToEd25519PubKey(privKey.PubKey().String())
		assert.ErrorIs(t, err, errInvalidPublicKeyType)
	})

	t.Run("valid public key type", func(t *testing.T) {
		t.Parallel()

		privKey := ed25519.GenPrivKey()

		_, err := Bech32ToEd25519PubKey(privKey.PubKey().String())
		assert.NoError(t, err)
	})
}

func TestLoadAuthKeysFile(t *testing.T) {
	t.Parallel()

	t.Run("non-existent file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(t.TempDir(), "non-existent")

		authKeysFile, err := LoadAuthKeysFile(filePath)
		require.Nil(t, authKeysFile)
		assert.Error(t, err)
	})

	t.Run("invalid file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(t.TempDir(), "invalid")

		os.WriteFile(filePath, []byte("invalid"), 0o600)

		authKeysFile, err := LoadAuthKeysFile(filePath)
		require.Nil(t, authKeysFile)
		assert.Error(t, err)
	})

	t.Run("valid file with invalid authorized keys", func(t *testing.T) {
		t.Parallel()

		privKey := ed25519.GenPrivKey()
		filePath := filepath.Join(t.TempDir(), "valid")

		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: privKey,
				PubKey:  privKey.PubKey().String(),
			},
			ClientAuthorizedKeys: []string{"invalid"},
		}

		jsonBytes, err := amino.MarshalJSONIndent(akf, "", "  ")
		require.NoError(t, err)
		os.WriteFile(filePath, jsonBytes, 0o600)

		authKeysFile, err := LoadAuthKeysFile(filePath)
		require.Nil(t, authKeysFile)
		assert.Error(t, err)
	})

	t.Run("valid file", func(t *testing.T) {
		t.Parallel()

		privKey := ed25519.GenPrivKey()
		filePath := filepath.Join(t.TempDir(), "valid")

		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: privKey,
				PubKey:  privKey.PubKey().String(),
			},
			ClientAuthorizedKeys: []string{privKey.PubKey().String()},
		}

		jsonBytes, err := amino.MarshalJSONIndent(akf, "", "  ")
		require.NoError(t, err)
		os.WriteFile(filePath, jsonBytes, 0o600)

		authKeysFile, err := LoadAuthKeysFile(filePath)
		require.NotNil(t, authKeysFile)
		assert.NoError(t, err)
	})
}

func TestAuthorizedKeys(t *testing.T) {
	t.Parallel()

	t.Run("valid authorized keys", func(t *testing.T) {
		t.Parallel()

		privKey := ed25519.GenPrivKey()
		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: privKey,
				PubKey:  privKey.PubKey().String(),
			},
			ClientAuthorizedKeys: []string{privKey.PubKey().String()},
		}

		filePath := filepath.Join(t.TempDir(), "valid")
		require.NoError(t, akf.Save(filePath))

		loaded, err := LoadAuthKeysFile(filePath)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		authorizedKeys := loaded.AuthorizedKeys()
		require.NotNil(t, authorizedKeys)
		require.Len(t, authorizedKeys, 1)
		assert.Equal(t, privKey.PubKey(), authorizedKeys[0])
	})
}

func TestGeneratePersistedAuthKeysFile(t *testing.T) {
	t.Parallel()

	t.Run("invalid parent dir", func(t *testing.T) {
		t.Parallel()

		// Create an unwriteable directory.
		basePath := filepath.Join(t.TempDir(), "unwriteable")
		require.NoError(t, os.MkdirAll(basePath, 0o000))
		_, err := GeneratePersistedAuthKeysFile(filepath.Join(basePath, "file"))
		require.Error(t, err)

		// Restore the permissions for cleanup.
		assert.NoError(t, os.Chmod(basePath, 0o700))
	})

	t.Run("valid auth keys file", func(t *testing.T) {
		t.Parallel()

		authKeysFile, err := GeneratePersistedAuthKeysFile(
			filepath.Join(t.TempDir(), "file"),
		)
		require.NotNil(t, authKeysFile)
		assert.NoError(t, err)
	})
}
