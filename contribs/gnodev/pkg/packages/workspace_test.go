package packages

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindWorkspace_ModFileInCwd(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnomod.toml"), []byte(""), 0o644))

	got := FindWorkspace(dir)
	assert.Equal(t, dir, got)
}

func TestFindWorkspace_WorkFileInAncestor(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))
	child := filepath.Join(root, "sub", "deeper")
	require.NoError(t, os.MkdirAll(child, 0o755))

	got := FindWorkspace(child)
	assert.Equal(t, root, got)
}

func TestFindWorkspace_None(t *testing.T) {
	dir := t.TempDir()
	assert.Equal(t, "", FindWorkspace(dir))
}

// A bare gnomod.toml in an ancestor (no gnowork.toml) must NOT be treated as
// the workspace root: gnovm's loader context honors gnomod.toml only in CWD,
// so handing it `ancestor/...` from a subdirectory crashes startup. Returning
// "" routes the caller to discovery mode instead.
func TestFindWorkspace_ModFileInAncestorIsNotWorkspace(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnomod.toml"), []byte(""), 0o644))
	child := filepath.Join(root, "sub", "deeper")
	require.NoError(t, os.MkdirAll(child, 0o755))

	assert.Equal(t, "", FindWorkspace(child))
}
