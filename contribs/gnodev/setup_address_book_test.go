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
// for testing the "name present, address differs" branch of ensureDevKey.
const otherMnemonic = "equip will roof matter pink blind book anxiety banner elbow sun young"

func newCaptureLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), &buf
}

func TestEnsureDevKey_EmptyKeybase(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	logger, buf := newCaptureLogger()

	cfg := &AppConfig{home: home}
	ensureDevKey(logger, cfg)

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	info, err := kb.GetByName(DevKeyName)
	require.NoError(t, err)
	assert.Equal(t, defaultDeployerAddress, info.GetAddress())

	assert.Contains(t, buf.String(), "dev key imported")
}

func TestEnsureDevKey_AlreadyPresentMatchingAddress(t *testing.T) {
	t.Parallel()
	home := t.TempDir()

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	_, err = kb.CreateAccount(DevKeyName, DefaultDeployerSeed, "", "", 0, 0)
	require.NoError(t, err)

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: home}
	ensureDevKey(logger, cfg)

	info, err := kb.GetByName(DevKeyName)
	require.NoError(t, err)
	assert.Equal(t, defaultDeployerAddress, info.GetAddress())

	logs := buf.String()
	assert.Contains(t, logs, "already present")
	assert.NotContains(t, logs, "dev key imported")
}

func TestEnsureDevKey_NamePresentConflictingAddress(t *testing.T) {
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
	ensureDevKey(logger, cfg)

	info, err := kb.GetByName(DevKeyName)
	require.NoError(t, err)
	assert.Equal(t, pre.GetAddress(), info.GetAddress(),
		"existing dev key entry must remain untouched")

	logs := buf.String()
	assert.Contains(t, logs, "different address")
	assert.Contains(t, logs, "not overwriting")
}

func TestEnsureDevKey_OptOut(t *testing.T) {
	t.Parallel()
	home := t.TempDir()

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: home, noDevKey: true}
	ensureDevKey(logger, cfg)

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	has, err := kb.HasByName(DevKeyName)
	require.NoError(t, err)
	assert.False(t, has, "-no-dev-key must not import the key")

	assert.Contains(t, buf.String(), "-no-dev-key")
}

func TestEnsureDevKey_NoHome(t *testing.T) {
	t.Parallel()

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: ""}
	ensureDevKey(logger, cfg)

	assert.Contains(t, buf.String(), "home not specified")
}

func TestEnsureDevKey_HomeMissing(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "does", "not", "exist")
	require.False(t, osm.DirExists(missing), "sanity: path must not exist")

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: missing}
	ensureDevKey(logger, cfg)

	assert.False(t, osm.DirExists(missing),
		"ensureDevKey must not materialize a missing -home")
	assert.Contains(t, buf.String(), "home directory does not exist")
}

func TestEnsureDevKey_DefaultHomeMissingIsCreated(t *testing.T) {
	// Not parallel: mutates GNOHOME via t.Setenv.
	fresh := filepath.Join(t.TempDir(), "fresh-install")
	t.Setenv("GNOHOME", fresh)
	require.Equal(t, fresh, gnoenv.HomeDir(),
		"sanity: GNOHOME must drive gnoenv.HomeDir()")
	require.False(t, osm.DirExists(fresh), "sanity: path must not exist yet")

	logger, buf := newCaptureLogger()
	cfg := &AppConfig{home: fresh}
	ensureDevKey(logger, cfg)

	assert.True(t, osm.DirExists(fresh),
		"default home must be materialized on first run")
	kb, err := keys.NewKeyBaseFromDir(fresh)
	require.NoError(t, err)
	info, err := kb.GetByName(DevKeyName)
	require.NoError(t, err)
	assert.Equal(t, defaultDeployerAddress, info.GetAddress())
	assert.Contains(t, buf.String(), "dev key imported")
}

func TestSetupAddressBook_AutoImportPutsDevKeyInBook(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	logger, _ := newCaptureLogger()

	book, err := setupAddressBook(logger, &AppConfig{home: home})
	require.NoError(t, err)

	names, ok := book.GetByAddress(defaultDeployerAddress)
	require.True(t, ok, "deployer address must be in the book")
	assert.Contains(t, names, DevKeyName,
		"deployer address must be resolvable under the dev name")
}

func TestSetupAddressBook_NoDevKeyFallsBackInMemory(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	logger, buf := newCaptureLogger()

	book, err := setupAddressBook(logger, &AppConfig{home: home, noDevKey: true})
	require.NoError(t, err)

	_, ok := book.GetByAddress(defaultDeployerAddress)
	require.True(t, ok, "deployer address must still be tracked in-memory")

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	has, err := kb.HasByName(DevKeyName)
	require.NoError(t, err)
	assert.False(t, has, "--no-dev-key must not import the key into the keybase")

	logs := buf.String()
	assert.Contains(t, logs, "tracked in-memory only")
	assert.NotContains(t, logs, DefaultDeployerSeed,
		"fallback log must not echo the mnemonic")
}

// The deployer seed already imported under another name (commonly test1)
// must be left untouched: gnodev detects the address is already signable and
// skips the import, rather than letting CreateAccount rename the entry to dev.
func TestEnsureDevKey_DeployerAddressUnderOtherNameIsPreserved(t *testing.T) {
	t.Parallel()
	home := t.TempDir()

	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	pre, err := kb.CreateAccount("test1", DefaultDeployerSeed, "", "", 0, 0)
	require.NoError(t, err)
	require.Equal(t, defaultDeployerAddress, pre.GetAddress(),
		"sanity: test1 must map to the deployer address")

	logger, buf := newCaptureLogger()
	ensureDevKey(logger, &AppConfig{home: home})

	hasTest1, err := kb.HasByName("test1")
	require.NoError(t, err)
	assert.True(t, hasTest1, "existing test1 entry must be preserved")
	hasDev, err := kb.HasByName(DevKeyName)
	require.NoError(t, err)
	assert.False(t, hasDev, "no second name must be added for an address already present")

	assert.Contains(t, buf.String(), "already present")
}

// A keybase that cannot be read (here: a regular file where the leveldb dir is
// expected) must not abort gnodev; ensureDevKey logs and returns.
func TestEnsureDevKey_BrokenKeybaseDegradesGracefully(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	// keys.NewKeyBaseFromDir opens <home>/data; a file there makes every
	// keybase read fail with a "not a directory" error.
	require.NoError(t, os.WriteFile(filepath.Join(home, "data"), []byte("x"), 0o600))

	logger, buf := newCaptureLogger()
	require.NotPanics(t, func() { ensureDevKey(logger, &AppConfig{home: home}) })

	assert.Contains(t, buf.String(), "dev key skipped")
}

// When the default home does not exist and cannot be created (unwritable
// parent), ensureDevKey skips rather than failing.
func TestEnsureDevKey_CannotCreateDefaultHome(t *testing.T) {
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
	ensureDevKey(logger, &AppConfig{home: fresh})

	assert.False(t, osm.DirExists(fresh), "must not create the home under an unwritable parent")
	assert.Contains(t, buf.String(), "cannot create default home")
}

// An existing but unwritable home makes keys.NewKeyBaseFromDir panic while
// creating its data dir; ensureDevKey recovers and skips instead of crashing.
func TestEnsureDevKey_UnwritableHomeDegradesGracefully(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	home := t.TempDir()
	require.NoError(t, os.Chmod(home, 0o500))
	t.Cleanup(func() { _ = os.Chmod(home, 0o700) })

	logger, buf := newCaptureLogger()
	require.NotPanics(t, func() { ensureDevKey(logger, &AppConfig{home: home}) })

	assert.Contains(t, buf.String(), "cannot open keybase")
}
