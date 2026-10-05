package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveScriptFiles(t *testing.T) {
	dir := t.TempDir()
	a := writeScriptFile(t, dir, "a.txtar", "validators: 1\n")
	b := writeScriptFile(t, dir, "b.txtar", "validators: 1\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.md"), []byte("ignored"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0o755))

	other := t.TempDir()
	c := writeScriptFile(t, other, "c.txtar", "validators: 2\n")

	t.Run("a directory yields its scripts and nothing else", func(t *testing.T) {
		got, err := ResolveScriptFiles([]string{dir})
		require.NoError(t, err)
		assert.Equal(t, []string{a, b}, got)
	})

	t.Run("named files are taken as given", func(t *testing.T) {
		got, err := ResolveScriptFiles([]string{b, a})
		require.NoError(t, err)
		assert.Equal(t, []string{b, a}, got, "discovery preserves the order it was asked for")
	})

	t.Run("directories and files mix", func(t *testing.T) {
		got, err := ResolveScriptFiles([]string{dir, c})
		require.NoError(t, err)
		assert.Equal(t, []string{a, b, c}, got)
	})

	t.Run("a missing path names itself", func(t *testing.T) {
		_, err := ResolveScriptFiles([]string{filepath.Join(dir, "absent.txtar")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absent.txtar")
	})

	t.Run("a named file that is not a script is a mistake, not a skip", func(t *testing.T) {
		_, err := ResolveScriptFiles([]string{filepath.Join(dir, "notes.md")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), ".txtar")
	})

	t.Run("a directory holding no scripts says so", func(t *testing.T) {
		empty := t.TempDir()
		_, err := ResolveScriptFiles([]string{empty})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no .txtar")
	})
}
