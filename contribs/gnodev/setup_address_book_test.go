package main

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCaptureLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), &buf
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

	// What the `I` key does once the import lands: the book gains the name it
	// had no key for at boot.
	t.Run("the I key names the address it imports", func(t *testing.T) {
		home := t.TempDir()
		logger, _ := newCaptureLogger()

		book, err := setupAddressBook(logger, &AppConfig{home: home})
		require.NoError(t, err)
		names, _ := book.GetByAddress(defaultDeployerAddress)
		require.Empty(t, names, "nameless until a key exists for it")

		created, err := importDevKey(home)
		require.NoError(t, err)
		require.True(t, created)
		book.Add(defaultDeployerAddress, DevKeyName)

		names, ok := book.GetByAddress(defaultDeployerAddress)
		require.True(t, ok)
		assert.Equal(t, []string{DevKeyName}, names)
	})
}
