package local

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"unicode"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	t.Parallel()

	// Test empty file path.
	fk := &FileKey{}
	require.Error(t, fk.save())

	// Test read only file path.
	fk.filePath = path.Join(t.TempDir(), "unwritable")
	file, err := os.OpenFile(fk.filePath, os.O_CREATE|os.O_RDONLY, 0444)
	require.NoError(t, err)
	defer file.Close()
	require.Error(t, fk.save())

	// Test regular file path.
	fk.filePath = path.Join(t.TempDir(), "writable")
	require.NoError(t, fk.save())
}

func generateRandomFileKey(t *testing.T) *FileKey {
	t.Helper()

	fk := &FileKey{}
	fk.PrivKey = ed25519.GenPrivKey()
	fk.PubKey = fk.PrivKey.PubKey()
	fk.Address = fk.PubKey.Address()
	fk.filePath = path.Join(t.TempDir(), fk.PubKey.String())

	return fk
}

func TestLoadFileKey(t *testing.T) {
	t.Parallel()

	// Test non-existent file path.
	fk, err := loadFileKey("non-existent")
	require.Nil(t, fk)
	require.Error(t, err)

	// Test invalid filekey.
	invalidFile := path.Join(t.TempDir(), "invalid")
	os.WriteFile(invalidFile, []byte(`{address:"invalid"}`), 0644)
	fk, err = loadFileKey(invalidFile)
	require.Nil(t, fk)
	require.Error(t, err)

	// Test valid filekey.
	fk = generateRandomFileKey(t)
	fk.save()
}

func TestFileKeyMarshalling(t *testing.T) {
	t.Parallel()

	// Generate a random file key.
	fk := generateRandomFileKey(t)
	pubBytes := [32]byte(fk.PubKey.(ed25519.PubKeyEd25519))
	privBytes := [64]byte(fk.PrivKey.(ed25519.PrivKeyEd25519))
	pubB64 := base64.StdEncoding.EncodeToString(pubBytes[:])
	privB64 := base64.StdEncoding.EncodeToString(privBytes[:])

	// Format the file key to JSON.
	json := fmt.Sprintf(`{
  "address": "%s",
  "pub_key": {
    "@type": "/tm.PubKeyEd25519",
    "value": "%s"
  },
  "priv_key": {
    "@type": "/tm.PrivKeyEd25519",
    "value": "%s"
  }
}`, fk.Address, pubB64, privB64)

	// Helper to make sure the JSON strings are comparable.
	removeWhitespaces := func(s string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, s)
	}

	// Marshal the file key to JSON.
	marshalled, err := amino.MarshalJSON(fk)
	require.NotNil(t, marshalled)
	require.NoError(t, err)

	// Make sure the JSON strings match.
	require.Equal(t, removeWhitespaces(json), removeWhitespaces(string(marshalled)))

	// Unmarshal the JSON into a file key.
	unmarshalled := FileKey{}
	err = amino.UnmarshalJSON([]byte(json), &unmarshalled)
	require.NoError(t, err)

	// Make sure the values match.
	require.Equal(t, fk.Address, unmarshalled.Address)
	require.Equal(t, fk.PrivKey, unmarshalled.PrivKey)
	require.Equal(t, fk.PubKey, unmarshalled.PubKey)
}
