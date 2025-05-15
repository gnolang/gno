package gnokey

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	keyName     = "key"
	keyPassword = "password"
)

func generateKeyBaseWithKey(t *testing.T) (string, keys.Keybase) {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), "keybase")

	// Create a new keybase.
	keyBase, _ := keys.NewKeyBaseFromDir(filePath)
	require.NotNil(t, keyBase)

	// Create a new key.
	err := keyBase.ImportPrivKey(keyName, ed25519.GenPrivKey(), keyPassword)
	require.NoError(t, err)

	return filePath, keyBase
}

func TestNewGnokeySigner(t *testing.T) {
	t.Parallel()

	t.Run("with unknown keyname", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "valid")

		signer, err := newGnokeySigner(
			&gnokeyFlags{home: filePath},
			"unknown",
			commands.NewTestIO(),
		)
		require.Nil(t, signer)
		assert.Error(t, err)
	})

	t.Run("invalid password then valid", func(t *testing.T) {
		t.Parallel()

		// Generate a keybase with a key.
		filePath, keybase := generateKeyBaseWithKey(t)
		defer keybase.CloseDB()

		// Create a stdin with the password.
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(fmt.Sprintf("%s\n%s\n", "invalid", keyPassword)))

		signer, err := newGnokeySigner(
			&gnokeyFlags{
				home:                  filePath,
				insecurePasswordStdin: true,
			},
			keyName,
			io,
		)
		require.NotNil(t, signer)
		assert.NoError(t, err)
	})

	t.Run("closed stdin", func(t *testing.T) {
		t.Parallel()

		// Generate a keybase with a key.
		filePath, keybase := generateKeyBaseWithKey(t)
		defer keybase.CloseDB()

		// Create a stdin with the password.
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(""))

		signer, err := newGnokeySigner(
			&gnokeyFlags{
				home:                  filePath,
				insecurePasswordStdin: true,
			},
			keyName,
			io,
		)
		require.Nil(t, signer)
		assert.Error(t, err)
	})

	t.Run("invalid key", func(t *testing.T) {
		t.Parallel()

		// Generate a keybase with a key.
		filePath, keybase := generateKeyBaseWithKey(t)
		defer keybase.CloseDB()

		// Create an invalid key.
		info, err := keybase.CreateOffline("offline", ed25519.GenPrivKey().PubKey())
		require.NotNil(t, info)
		require.NoError(t, err)

		// Create a stdin with the password.
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(""))

		signer, err := newGnokeySigner(
			&gnokeyFlags{
				home:                  filePath,
				insecurePasswordStdin: true,
			},
			"offline",
			io,
		)
		require.Nil(t, signer)
		assert.Error(t, err)
	})

	t.Run("valid key", func(t *testing.T) {
		t.Parallel()

		// Generate a keybase with a key.
		filePath, keybase := generateKeyBaseWithKey(t)
		defer keybase.CloseDB()

		// Create a stdin with the password.
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(fmt.Sprintf("%s\n", keyPassword)))

		signer, err := newGnokeySigner(
			&gnokeyFlags{
				home:                  filePath,
				insecurePasswordStdin: true,
			},
			keyName,
			io,
		)
		require.NotNil(t, signer)
		require.NoError(t, err)

		// Test Signer interface.
		pubKey := signer.PubKey()
		require.NotNil(t, pubKey)

		signature, err := signer.Sign([]byte("test"))
		require.NotNil(t, signature)
		require.NoError(t, err)

		assert.NoError(t, signer.Close())
	})
}
