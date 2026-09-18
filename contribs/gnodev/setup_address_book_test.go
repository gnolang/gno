package main

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// otherMnemonic is a valid BIP-39 phrase distinct from DefaultDeployerSeed,
// for testing the "name present, address differs" branch of importDevKey.
const otherMnemonic = "equip will roof matter pink blind book anxiety banner elbow sun young"

func newCaptureLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), &buf
}

func TestImportDevKey_EmptyKeybase(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	logger, buf := newCaptureLogger()

	cfg := &AppConfig{home: home}
	importDevKey(logger, cfg.home)

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	info, err := kb.GetByName(DevKeyName)
	require.NoError(t, err)
	assert.Equal(t, defaultDeployerAddress, info.GetAddress())

	assert.Contains(t, buf.String(), "dev key imported")
}

func TestImportDevKey_AlreadyPresentMatchingAddress(t *testing.T) {
	t.Parallel()
	home := t.TempDir()

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	_, err = kb.CreateAccount(DevKeyName, DefaultDeployerSeed, "", "", 0, 0)
	require.NoError(t, err)

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: home}
	importDevKey(logger, cfg.home)

	info, err := kb.GetByName(DevKeyName)
	require.NoError(t, err)
	assert.Equal(t, defaultDeployerAddress, info.GetAddress())

	logs := buf.String()
	assert.Contains(t, logs, "already present")
	assert.NotContains(t, logs, "dev key imported")
}

func TestImportDevKey_NamePresentConflictingAddress(t *testing.T) {
	t.Parallel()
	home := t.TempDir()

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	pre, err := kb.CreateAccount(DevKeyName, otherMnemonic, "", "", 0, 0)
	require.NoError(t, err)
	require.NotEqual(t, defaultDeployerAddress, pre.GetAddress(),
		"sanity: chosen mnemonic must derive a different address than the deployer")

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: home}
	importDevKey(logger, cfg.home)

	info, err := kb.GetByName(DevKeyName)
	require.NoError(t, err)
	assert.Equal(t, pre.GetAddress(), info.GetAddress(),
		"existing dev key entry must remain untouched")

	logs := buf.String()
	assert.Contains(t, logs, "different address")
	assert.Contains(t, logs, "not overwriting")
}

// The import is opt-in: a plain boot writes nothing and says how to ask.
func TestSetupAddressBook_DefaultWritesNothing(t *testing.T) {
	t.Parallel()
	home := t.TempDir()

	logger, buf := newCaptureLogger()
	_, err := setupAddressBook(logger, &AppConfig{home: home})
	require.NoError(t, err)

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	has, err := kb.HasByName(DevKeyName)
	require.NoError(t, err)
	assert.False(t, has, "a plain boot must not write to the keybase")

	assert.Contains(t, buf.String(), "-import-dev-key")
}

func TestImportDevKey_NoHome(t *testing.T) {
	t.Parallel()

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: ""}
	importDevKey(logger, cfg.home)

	assert.Contains(t, buf.String(), "home not specified")
}

func TestImportDevKey_HomeMissing(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "does", "not", "exist")
	require.False(t, osm.DirExists(missing), "sanity: path must not exist")

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: missing}
	importDevKey(logger, cfg.home)

	assert.False(t, osm.DirExists(missing),
		"importDevKey must not materialize a missing -home")
	assert.Contains(t, buf.String(), "home directory does not exist")
}

func TestImportDevKey_DefaultHomeMissingIsCreated(t *testing.T) {
	// Not parallel: mutates GNOHOME via t.Setenv.
	fresh := filepath.Join(t.TempDir(), "fresh-install")
	t.Setenv("GNOHOME", fresh)
	require.Equal(t, fresh, gnoenv.HomeDir(),
		"sanity: GNOHOME must drive gnoenv.HomeDir()")
	require.False(t, osm.DirExists(fresh), "sanity: path must not exist yet")

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: fresh}
	importDevKey(logger, cfg.home)

	assert.True(t, osm.DirExists(fresh),
		"default home must be materialized on first run")
	kb, err := keys.NewKeyBaseFromDir(fresh)
	require.NoError(t, err)
	info, err := kb.GetByName(DevKeyName)
	require.NoError(t, err)
	assert.Equal(t, defaultDeployerAddress, info.GetAddress())
	assert.Contains(t, buf.String(), "dev key imported")
}

func TestSetupAddressBook_ImportFlagPutsDevKeyInBook(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	logger, _ := newCaptureLogger()

	book, err := setupAddressBook(logger, &AppConfig{home: home, importDevKey: true})
	require.NoError(t, err)

	names, ok := book.GetByAddress(defaultDeployerAddress)
	require.True(t, ok, "deployer address must be in the book")
	assert.Contains(t, names, DevKeyName,
		"deployer address must be resolvable under the dev name")
}

func TestSetupAddressBook_FallsBackInMemory(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	logger, buf := newCaptureLogger()

	book, err := setupAddressBook(logger, &AppConfig{home: home})
	require.NoError(t, err)

	_, ok := book.GetByAddress(defaultDeployerAddress)
	require.True(t, ok, "deployer address must still be tracked in-memory")

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	has, err := kb.HasByName(DevKeyName)
	require.NoError(t, err)
	assert.False(t, has, "a plain boot must not import the key into the keybase")

	logs := buf.String()
	assert.Contains(t, logs, "tracked in-memory only")
	assert.NotContains(t, logs, DefaultDeployerSeed,
		"fallback log must not echo the mnemonic")
}

// What the `I` keypress does: import, then replace the placeholder name
// gnodev invented, so the accounts panel shows one row under the real name.
func TestImportDevKey_KeypressRenamesPlaceholder(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	logger, _ := newCaptureLogger()

	book, err := setupAddressBook(logger, &AppConfig{home: home})
	require.NoError(t, err)
	names, ok := book.GetByAddress(defaultDeployerAddress)
	require.True(t, ok)
	require.NotContains(t, names, DevKeyName)

	require.True(t, importDevKey(logger, home))
	book.Rename(defaultDeployerAddress, DevKeyName)

	names, ok = book.GetByAddress(defaultDeployerAddress)
	require.True(t, ok)
	assert.Equal(t, []string{DevKeyName}, names)
}

// The deployer seed already imported under another name (commonly test1)
// must be left untouched: gnodev detects the address is already signable and
// skips the import, rather than letting CreateAccount rename the entry to devtest.
func TestImportDevKey_DeployerAddressUnderOtherNameIsPreserved(t *testing.T) {
	t.Parallel()
	home := t.TempDir()

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	pre, err := kb.CreateAccount("test1", DefaultDeployerSeed, "", "", 0, 0)
	require.NoError(t, err)
	require.Equal(t, defaultDeployerAddress, pre.GetAddress(),
		"sanity: test1 must map to the deployer address")

	logger, buf := newCaptureLogger()
	importDevKey(logger, home)

	hasTest1, err := kb.HasByName("test1")
	require.NoError(t, err)
	assert.True(t, hasTest1, "existing test1 entry must be preserved")
	hasDev, err := kb.HasByName(DevKeyName)
	require.NoError(t, err)
	assert.False(t, hasDev, "no second name must be added for an address already present")

	assert.Contains(t, buf.String(), "already present")
}

// A keybase that cannot be read (here: a regular file where the leveldb dir is
// expected) must not abort gnodev; importDevKey logs and returns.
func TestImportDevKey_BrokenKeybaseDegradesGracefully(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	// keys.NewKeyBaseFromDir opens <home>/data; a file there makes every
	// keybase read fail with a "not a directory" error.
	require.NoError(t, os.WriteFile(filepath.Join(home, "data"), []byte("x"), 0o600))

	logger, buf := newCaptureLogger()
	require.NotPanics(t, func() { importDevKey(logger, home) })

	assert.Contains(t, buf.String(), "dev key skipped")
}

// When the default home does not exist and cannot be created (unwritable
// parent), importDevKey skips rather than failing.
func TestImportDevKey_CannotCreateDefaultHome(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	parent := t.TempDir()
	require.NoError(t, os.Chmod(parent, 0o500))
	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })

	fresh := filepath.Join(parent, "gno")
	t.Setenv("GNOHOME", fresh)
	require.Equal(t, fresh, gnoenv.HomeDir(), "sanity: GNOHOME must drive gnoenv.HomeDir()")
	require.False(t, osm.DirExists(fresh), "sanity: path must not exist")

	logger, buf := newCaptureLogger()
	importDevKey(logger, fresh)

	assert.False(t, osm.DirExists(fresh), "must not create the home under an unwritable parent")
	assert.Contains(t, buf.String(), "cannot create default home")
}

// An existing but unwritable home makes keys.NewKeyBaseFromDir panic while
// creating its data dir; importDevKey recovers and skips instead of crashing.
func TestImportDevKey_UnwritableHomeDegradesGracefully(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	home := t.TempDir()
	require.NoError(t, os.Chmod(home, 0o500))
	t.Cleanup(func() { _ = os.Chmod(home, 0o700) })

	logger, buf := newCaptureLogger()
	require.NotPanics(t, func() { importDevKey(logger, home) })

	assert.Contains(t, buf.String(), "cannot open keybase")
}
