package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// keybaseWith returns a home whose keybase holds one key, derived from the
// well-known deployer seed at index: 0 is the deployer address itself, and any
// other index is an unrelated address.
func keybaseWith(t *testing.T, name string, index uint32) string {
	t.Helper()
	home := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	_, err = kb.CreateAccount(name, DefaultDeployerSeed, "", "", 0, index)
	require.NoError(t, err)
	return home
}

func keyAddress(t *testing.T, home, name string) (addr string, found bool) {
	t.Helper()
	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	info, err := kb.GetByName(name)
	if err != nil {
		return "", false
	}
	return info.GetAddress().String(), true
}

func TestImportDevKey(t *testing.T) {
	deployer := defaultDeployerAddress.String()

	t.Run("an empty keybase takes the key", func(t *testing.T) {
		home := t.TempDir()

		created, err := importDevKey(home)
		require.NoError(t, err)
		assert.True(t, created)

		addr, found := keyAddress(t, home, DevKeyName)
		require.True(t, found)
		assert.Equal(t, deployer, addr)
	})

	// The keybase holds one name per address, so importing here would delete
	// the user's own entry for it.
	t.Run("an address already held keeps its own name", func(t *testing.T) {
		home := keybaseWith(t, DefaultDeployerName, 0)

		created, err := importDevKey(home)
		assert.False(t, created)
		assert.ErrorIs(t, err, errDevKeyPresent)

		_, found := keyAddress(t, home, DefaultDeployerName)
		assert.True(t, found, "the user's own name survives")
		_, found = keyAddress(t, home, DevKeyName)
		assert.False(t, found, "no second name for one address")
	})

	t.Run("a name taken by another address is left alone", func(t *testing.T) {
		home := keybaseWith(t, DevKeyName, 1)

		created, err := importDevKey(home)
		assert.False(t, created)
		assert.ErrorContains(t, err, "already names")

		addr, found := keyAddress(t, home, DevKeyName)
		require.True(t, found)
		assert.NotEqual(t, deployer, addr)
	})

	t.Run("no home to write to", func(t *testing.T) {
		created, err := importDevKey("")
		assert.False(t, created)
		assert.ErrorContains(t, err, "no home directory")
	})

	t.Run("a missing -home is never materialized", func(t *testing.T) {
		home := filepath.Join(t.TempDir(), "does", "not", "exist")

		created, err := importDevKey(home)
		assert.False(t, created)
		assert.ErrorContains(t, err, "does not exist")
		assert.False(t, osm.DirExists(home))
	})

	t.Run("a missing default home is created", func(t *testing.T) {
		home := filepath.Join(t.TempDir(), "fresh-install")
		t.Setenv("GNOHOME", home)

		created, err := importDevKey(home)
		require.NoError(t, err)
		assert.True(t, created)
		assert.True(t, osm.DirExists(home))
	})

	// A file where the leveldb directory belongs: every read fails.
	t.Run("an unreadable keybase boots anyway", func(t *testing.T) {
		home := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(home, "data"), []byte("x"), 0o600))

		created, err := importDevKey(home)
		assert.False(t, created)
		assert.Error(t, err)
	})

	t.Run("an unwritable home boots anyway", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}
		home := t.TempDir()
		require.NoError(t, os.Chmod(home, 0o500))
		t.Cleanup(func() { _ = os.Chmod(home, 0o700) })

		created, err := importDevKey(home)
		assert.False(t, created)
		assert.ErrorContains(t, err, "keybase")
	})
}

func TestDevKeySignable(t *testing.T) {
	t.Run("a home with no keybase in it", func(t *testing.T) {
		assert.False(t, devKeySignable(t.TempDir()))
	})

	t.Run("no home at all", func(t *testing.T) {
		assert.False(t, devKeySignable(""))
	})

	t.Run("a keybase holding another key", func(t *testing.T) {
		assert.False(t, devKeySignable(keybaseWith(t, "mine", 1)))
	})

	t.Run("a keybase holding the deployer", func(t *testing.T) {
		assert.True(t, devKeySignable(keybaseWith(t, DefaultDeployerName, 0)))
	})
}

func TestAskDevKey(t *testing.T) {
	ask := func(t *testing.T, home, answer string) string {
		t.Helper()
		var out bytes.Buffer
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(answer))
		io.SetErr(commands.WriteNopCloser(&out))

		askDevKey(io, home)
		return out.String()
	}

	t.Run("y adds the key", func(t *testing.T) {
		home := t.TempDir()

		out := ask(t, home, "y\n")
		assert.Contains(t, out, "test account")
		assert.Contains(t, out, "dev key added")

		addr, found := keyAddress(t, home, DevKeyName)
		require.True(t, found)
		assert.Equal(t, defaultDeployerAddress.String(), addr)
	})

	t.Run("yes adds the key", func(t *testing.T) {
		home := t.TempDir()

		ask(t, home, "yes\n")

		_, found := keyAddress(t, home, DevKeyName)
		assert.True(t, found)
	})

	for _, answer := range []string{"\n", "n\n", "no\n", ""} {
		t.Run("declining on "+strings.TrimSuffix(answer, "\n")+" writes nothing", func(t *testing.T) {
			home := t.TempDir()

			out := ask(t, home, answer)
			assert.Contains(t, out, "test account")

			_, found := keyAddress(t, home, DevKeyName)
			assert.False(t, found)
		})
	}

	t.Run("a keybase that already signs is not asked", func(t *testing.T) {
		out := ask(t, keybaseWith(t, DefaultDeployerName, 0), "y\n")
		assert.Empty(t, out)
	})

	t.Run("no home is not asked", func(t *testing.T) {
		assert.Empty(t, ask(t, "", "y\n"))
	})
}
