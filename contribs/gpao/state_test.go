package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStateRoundTrips: the cursor a run records is the cursor the next run
// reads. The whole feature is this sentence.
func TestStateRoundTrips(t *testing.T) {
	dir := t.TempDir()

	s, err := openStateStore(dir, "test-chain")
	require.NoError(t, err)
	require.Equal(t, noCursor, s.lastVerifiedHeight(), "a fresh data dir has no cursor")
	require.NoError(t, s.setLastVerifiedHeight(42))

	reopened, err := openStateStore(dir, "test-chain")
	require.NoError(t, err)
	assert.Equal(t, int64(42), reopened.lastVerifiedHeight())
}

// TestStateCreatesItsDirectory: --data-dir names where state should live, not
// where it already does, so a first run must not have to be prepared by hand --
// including under a parent that does not exist either.
func TestStateCreatesItsDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "gpao")

	s, err := openStateStore(dir, "test-chain")
	require.NoError(t, err)
	require.NoError(t, s.setLastVerifiedHeight(7))

	assert.FileExists(t, filepath.Join(dir, stateFileName))
}

// TestStateRefusesAnotherChainsCursor pins the guard that makes the default
// --data-dir safe. It sits under $GNOHOME, so the same directory will be reused
// against a second chain sooner or later, and a height from one chain names
// nothing on another.
func TestStateRefusesAnotherChainsCursor(t *testing.T) {
	dir := t.TempDir()

	s, err := openStateStore(dir, "portal-loop")
	require.NoError(t, err)
	require.NoError(t, s.setLastVerifiedHeight(4218))

	_, err = openStateStore(dir, "test3")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "portal-loop")
	assert.Contains(t, err.Error(), "test3",
		"the error has to name both chains, or it does not say what is wrong")
	assert.Contains(t, err.Error(), "--start-height",
		"and it has to name a way out")
}

// TestStateRefusesAnUndecodableFile: writes are atomic, so an undecodable file
// is not a torn write and starting over silently would discard a cursor that
// something else is confused about. It fails, and says what would fix it --
// otherwise the daemon just refuses to boot with no hint.
func TestStateRefusesAnUndecodableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, stateFileName)
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0o600))

	_, err := openStateStore(dir, "test-chain")
	require.Error(t, err)
	assert.Contains(t, err.Error(), path, "the error has to name the file")
	assert.Contains(t, err.Error(), "delete it")
}

// TestStateWriteIsMonotone: the writer is one goroutine draining a FIFO, so a
// rewind cannot happen today. Refused here anyway, so that a second writer
// added later cannot silently undo progress -- the failure that would look like
// an oracle re-verifying blocks forever.
func TestStateWriteIsMonotone(t *testing.T) {
	dir := t.TempDir()

	s, err := openStateStore(dir, "test-chain")
	require.NoError(t, err)
	require.NoError(t, s.setLastVerifiedHeight(100))
	require.NoError(t, s.setLastVerifiedHeight(50))

	assert.Equal(t, int64(100), s.lastVerifiedHeight(), "the cursor may not go backwards")

	// reset is the exception, and the only one: -start-height has to be able to
	// contradict a recorded height, or a height stored in error could never be
	// revisited.
	require.NoError(t, s.reset(50))
	assert.Equal(t, int64(50), s.lastVerifiedHeight())
}

// TestStateFileShape pins the on-disk field names. They are an operator
// interface -- something will grep this file -- so renaming one is a breaking
// change and should fail here first.
func TestStateFileShape(t *testing.T) {
	dir := t.TempDir()
	s, err := openStateStore(dir, "test-chain")
	require.NoError(t, err)
	require.NoError(t, s.setLastVerifiedHeight(9))

	raw, err := os.ReadFile(filepath.Join(dir, stateFileName))
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, "test-chain", decoded["chain_id"])
	assert.Equal(t, float64(9), decoded["last_verified_height"])

	// Indented, and the space after the colon with it. A scenario greps
	// `"last_verified_height": 0` to assert the cursor moved, so switching to
	// json.Marshal would not fail a test -- it would turn that negation into a
	// match on nothing and quietly stop checking anything.
	assert.Contains(t, string(raw), `"last_verified_height": 9`)
}

// TestStateSeparatesAZeroCursorFromNone: `-start-height 1` legitimately records
// "verified through 0". Reading that back as "nothing recorded" would let the
// next bare restart jump to the tip and silently abandon the replay the
// operator asked for, which is the failure the rewrite-on-override exists to
// prevent.
func TestStateSeparatesAZeroCursorFromNone(t *testing.T) {
	dir := t.TempDir()

	fresh, err := openStateStore(dir, "test-chain")
	require.NoError(t, err)
	require.Equal(t, noCursor, fresh.lastVerifiedHeight())

	require.NoError(t, fresh.reset(0))

	reopened, err := openStateStore(dir, "test-chain")
	require.NoError(t, err)
	assert.Zero(t, reopened.lastVerifiedHeight(),
		"a recorded 0 is not the absence of a cursor")
}

// TestStateRefusesAFileWithNoChainID: every file this daemon writes names its
// chain, so one that does not is a file something else produced. Adopting it
// for whatever chain happens to be asking is what the guard exists to stop.
func TestStateRefusesAFileWithNoChainID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, stateFileName)
	require.NoError(t, os.WriteFile(path, []byte(`{"last_verified_height": 99}`), 0o600))

	_, err := openStateStore(dir, "test-chain")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test-chain")
}
