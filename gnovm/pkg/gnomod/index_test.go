package gnomod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeModule writes a gnomod.toml declaring modPath at root/dir.
func writeModule(t *testing.T, root, dir, modPath string) string {
	t.Helper()

	full := filepath.Join(root, filepath.FromSlash(dir))
	require.NoError(t, os.MkdirAll(full, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(full, "gnomod.toml"),
		[]byte("module = \""+modPath+"\"\ngno = \"0.9\"\n"),
		0o644,
	))
	return full
}

func TestIndexDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	avl := writeModule(t, root, "gno.land/p/nt/avl", "gno.land/p/nt/avl/v0")
	minisocial := writeModule(t, root, "gno.land/r/docs/minisocial/v1", "gno.land/r/docs/minisocial/v1")
	shadowed := writeModule(t, root, "quarantined/gno.land/p/nt/avl", "gno.land/p/nt/avl/v0")

	index, err := IndexDir(root)
	require.NoError(t, err)

	// A directory that does not repeat the version of its module path is
	// indexed at its module path, not at its directory.
	assert.Equal(t, avl, index["gno.land/p/nt/avl/v0"])
	assert.NotContains(t, index, "gno.land/p/nt/avl")
	// A directory that does repeat it is indexed all the same.
	assert.Equal(t, minisocial, index["gno.land/r/docs/minisocial/v1"])
	// The first in lexical order wins a duplicate module path.
	assert.NotEqual(t, shadowed, index["gno.land/p/nt/avl/v0"])
}

func TestIndexDirMissingRoot(t *testing.T) {
	t.Parallel()

	_, err := IndexDir(filepath.Join(t.TempDir(), "absent"))
	assert.Error(t, err)
}

func TestFindModuleDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	avl := writeModule(t, root, "gno.land/p/nt/avl", "gno.land/p/nt/avl/v0")
	minisocial := writeModule(t, root, "gno.land/r/docs/minisocial/v1", "gno.land/r/docs/minisocial/v1")

	// Not a module: it can only be identified by where it sits.
	plain := filepath.Join(root, "gno.land", "p", "nt", "plain")
	require.NoError(t, os.MkdirAll(plain, 0o755))

	for _, tc := range []struct {
		name    string
		modPath string
		want    string
	}{
		{"version only in the module line", "gno.land/p/nt/avl/v0", avl},
		{"version in the directory too", "gno.land/r/docs/minisocial/v1", minisocial},
		{"directory without a gnomod.toml", "gno.land/p/nt/plain", plain},
		{"directory declaring another module", "gno.land/p/nt/avl", ""},
		{"no such module", "gno.land/p/nt/absent/v0", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, FindModuleDir(root, tc.modPath))
		})
	}
}

func TestIndexDirCached(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeModule(t, root, "gno.land/p/nt/avl", "gno.land/p/nt/avl/v0")

	cached, err := IndexDirCached(root)
	require.NoError(t, err)
	require.Len(t, cached, 1)

	writeModule(t, root, "gno.land/p/nt/ufmt", "gno.land/p/nt/ufmt/v0")

	// The walk happens once, and a trailing separator is the same root.
	again, err := IndexDirCached(root + string(filepath.Separator))
	require.NoError(t, err)
	assert.Len(t, again, 1)

	fresh, err := IndexDir(root)
	require.NoError(t, err)
	assert.Len(t, fresh, 2)
}
