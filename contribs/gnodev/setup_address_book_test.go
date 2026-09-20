package main

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCaptureLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), &buf
}

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

	for _, tc := range []struct {
		name string
		home func(t *testing.T) string
		want bool
		log  string
		then func(t *testing.T, home string)
	}{
		{
			name: "empty keybase takes the key",
			home: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			want: true,
			log:  "dev key imported",
			then: func(t *testing.T, home string) {
				t.Helper()
				addr, found := keyAddress(t, home, DevKeyName)
				require.True(t, found)
				assert.Equal(t, deployer, addr)
			},
		},
		{
			// The keybase holds one name per address, so importing here would
			// delete the user's own entry for it.
			name: "address already held under another name keeps that name",
			home: func(t *testing.T) string {
				t.Helper()
				return keybaseWith(t, DefaultDeployerName, 0)
			},
			want: true,
			log:  "already present",
			then: func(t *testing.T, home string) {
				t.Helper()
				_, found := keyAddress(t, home, DefaultDeployerName)
				assert.True(t, found, "the user's own name must survive")
				_, found = keyAddress(t, home, DevKeyName)
				assert.False(t, found, "no second name for one address")
			},
		},
		{
			name: "name taken by another address is left alone",
			home: func(t *testing.T) string {
				t.Helper()
				return keybaseWith(t, DevKeyName, 1)
			},
			want: false,
			log:  "not overwriting",
			then: func(t *testing.T, home string) {
				t.Helper()
				addr, found := keyAddress(t, home, DevKeyName)
				require.True(t, found)
				assert.NotEqual(t, deployer, addr)
			},
		},
		{
			name: "no home to write to",
			home: func(t *testing.T) string {
				t.Helper()
				return ""
			},
			want: false,
			log:  "home not specified",
		},
		{
			name: "a missing -home is never materialized",
			home: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "does", "not", "exist")
			},
			want: false,
			log:  "home directory does not exist",
			then: func(t *testing.T, home string) {
				t.Helper()
				assert.False(t, osm.DirExists(home))
			},
		},
		{
			name: "a missing default home is created",
			home: func(t *testing.T) string {
				t.Helper()
				fresh := filepath.Join(t.TempDir(), "fresh-install")
				t.Setenv("GNOHOME", fresh)
				return fresh
			},
			want: true,
			log:  "dev key imported",
			then: func(t *testing.T, home string) {
				t.Helper()
				require.True(t, osm.DirExists(home))
				addr, found := keyAddress(t, home, DevKeyName)
				require.True(t, found)
				assert.Equal(t, deployer, addr)
			},
		},
		{
			// A file where the leveldb directory belongs: every read fails.
			name: "unreadable keybase boots anyway",
			home: func(t *testing.T) string {
				t.Helper()
				home := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(home, "data"), []byte("x"), 0o600))
				return home
			},
			want: false,
			log:  "dev key skipped",
		},
		{
			name: "unwritable home boots anyway",
			home: func(t *testing.T) string {
				t.Helper()
				if os.Geteuid() == 0 {
					t.Skip("root bypasses directory permissions")
				}
				home := t.TempDir()
				require.NoError(t, os.Chmod(home, 0o500))
				t.Cleanup(func() { _ = os.Chmod(home, 0o700) })
				return home
			},
			want: false,
			log:  "cannot open keybase",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			home := tc.home(t)
			logger, buf := newCaptureLogger()

			var got bool
			require.NotPanics(t, func() { got = importDevKey(logger, home) })

			assert.Equal(t, tc.want, got)
			assert.Contains(t, buf.String(), tc.log)
			if tc.then != nil {
				tc.then(t, home)
			}
		})
	}
}

func TestDevKeySignable(t *testing.T) {
	t.Run("a home with no keybase in it", func(t *testing.T) {
		home := t.TempDir()
		assert.False(t, devKeySignable(home))
		assert.NoDirExists(t, filepath.Join(home, "data"), "the probe creates nothing")
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

func TestSetupAddressBook(t *testing.T) {
	t.Run("a plain boot writes nothing", func(t *testing.T) {
		home := t.TempDir()
		logger, buf := newCaptureLogger()

		book, err := setupAddressBook(logger, &AppConfig{home: home})
		require.NoError(t, err)

		_, ok := book.GetByAddress(defaultDeployerAddress)
		assert.True(t, ok, "the address is still premined")
		_, found := keyAddress(t, home, DevKeyName)
		assert.False(t, found, "nothing written to the keybase")

		logs := buf.String()
		assert.Contains(t, logs, "cannot sign with it")
		assert.NotContains(t, logs, DefaultDeployerSeed, "the mnemonic never reaches the log")
	})

	t.Run("an imported key names the address in the book", func(t *testing.T) {
		home := keybaseWith(t, DevKeyName, 0)
		logger, _ := newCaptureLogger()

		book, err := setupAddressBook(logger, &AppConfig{home: home})
		require.NoError(t, err)

		names, ok := book.GetByAddress(defaultDeployerAddress)
		require.True(t, ok)
		assert.Contains(t, names, DevKeyName)
	})

	t.Run("the I key names the address it imports", func(t *testing.T) {
		home := t.TempDir()
		logger, _ := newCaptureLogger()

		book, err := setupAddressBook(logger, &AppConfig{home: home})
		require.NoError(t, err)
		names, _ := book.GetByAddress(defaultDeployerAddress)
		require.Empty(t, names, "nameless until a key exists for it")

		require.True(t, importDevKey(logger, home))
		require.NoError(t, book.ImportKeybase(home))

		names, ok := book.GetByAddress(defaultDeployerAddress)
		require.True(t, ok)
		assert.Equal(t, []string{DevKeyName}, names)
	})
}
