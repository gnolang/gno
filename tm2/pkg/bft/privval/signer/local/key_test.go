package local

import (
	"os"
	"path"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid FileKey", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey("filePath")

		require.NoError(t, fk.validate())
	})

	t.Run("invalid private key", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey("filePath")
		fk.PrivKey = nil

		require.ErrorIs(t, fk.validate(), errInvalidPrivateKey)
	})

	t.Run("public key mismatch", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey("filePath")
		fk.PubKey = nil

		require.ErrorIs(t, fk.validate(), errPublicKeyMismatch)
	})

	t.Run("address mismatch", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey("filePath")
		fk.Address = crypto.Address{} // zero address

		require.ErrorIs(t, fk.validate(), errAddressMismatch)
	})

	t.Run("empty filepath", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey("")

		require.ErrorIs(t, fk.validate(), errFilePathNotSet)
	})
}

func TestSave(t *testing.T) {
	t.Parallel()

	t.Run("empty file path", func(t *testing.T) {
		t.Parallel()

		fk := GenerateFileKey("")
		require.Error(t, fk.save())
	})

	t.Run("read-only file path", func(t *testing.T) {
		t.Parallel()

		// Create a read-only file.
		filePath := path.Join(t.TempDir(), "unwritable")
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0444)
		require.NoError(t, err)
		defer file.Close()

		fk := GenerateFileKey(filePath)
		require.Error(t, fk.save())
	})

	t.Run("read-write file path", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "writable")
		fk := GenerateFileKey(filePath)
		require.NoError(t, fk.save())
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
		require.Equal(t, fk.PrivKey, loaded.PrivKey)
		require.Equal(t, fk.PubKey, loaded.PubKey)
		require.Equal(t, fk.Address, loaded.Address)
		require.Equal(t, fk.filePath, loaded.filePath)
	})

	t.Run("non-existent file path", func(t *testing.T) {
		t.Parallel()

		fk, err := LoadFileKey("non-existent")
		require.Nil(t, fk)
		require.Error(t, err)
	})

	t.Run("invalid file key", func(t *testing.T) {
		t.Parallel()

		// Create a file with invalid FileKey JSON.
		filePath := path.Join(t.TempDir(), "invalid")
		os.WriteFile(filePath, []byte(`{address:"invalid"}`), 0644)

		fk, err := LoadFileKey(filePath)
		require.Nil(t, fk)
		require.Error(t, err)
	})
}

// func TestFileKeyMarshalling(t *testing.T) {
// 	t.Parallel()
//
// 	// Generate a random file key.
// 	fk := GenerateFileKey("")
// 	pubBytes := [32]byte(fk.PubKey.(ed25519.PubKeyEd25519))
// 	privBytes := [64]byte(fk.PrivKey.(ed25519.PrivKeyEd25519))
// 	pubB64 := base64.StdEncoding.EncodeToString(pubBytes[:])
// 	privB64 := base64.StdEncoding.EncodeToString(privBytes[:])
//
// 	// Format the file key to JSON.
// 	json := fmt.Sprintf(`{
//   "address": "%s",
//   "pub_key": {
//     "@type": "/tm.PubKeyEd25519",
//     "value": "%s"
//   },
//   "priv_key": {
//     "@type": "/tm.PrivKeyEd25519",
//     "value": "%s"
//   }
// }`, fk.Address, pubB64, privB64)
//
// 	// Helper to make sure the JSON strings are comparable.
// 	removeWhitespaces := func(s string) string {
// 		return strings.Map(func(r rune) rune {
// 			if unicode.IsSpace(r) {
// 				return -1
// 			}
// 			return r
// 		}, s)
// 	}
//
// 	// Marshal the file key to JSON.
// 	marshalled, err := amino.MarshalJSON(fk)
// 	require.NotNil(t, marshalled)
// 	require.NoError(t, err)
//
// 	// Make sure the JSON strings match.
// 	require.Equal(t, removeWhitespaces(json), removeWhitespaces(string(marshalled)))
//
// 	// Unmarshal the JSON into a file key.
// 	unmarshalled := FileKey{}
// 	err = amino.UnmarshalJSON([]byte(json), &unmarshalled)
// 	require.NoError(t, err)
//
// 	// Make sure the values match.
// 	require.Equal(t, fk.Address, unmarshalled.Address)
// 	require.Equal(t, fk.PrivKey, unmarshalled.PrivKey)
// 	require.Equal(t, fk.PubKey, unmarshalled.PubKey)
// }
