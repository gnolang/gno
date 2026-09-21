package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackage_ToMemPackage_FS(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnomod.toml"),
		fmt.Appendf(nil, "module = %q\n", "gno.land/p/demo/foo"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "foo.gno"),
		[]byte("package foo\n"), 0o644))

	p := &Package{ImportPath: "gno.land/p/demo/foo", Dir: dir, Kind: KindFS}
	mp, err := p.ToMemPackage()
	require.NoError(t, err)
	assert.Equal(t, "gno.land/p/demo/foo", mp.Path)
	assert.Equal(t, "foo", mp.Name)
}

func TestPackage_ToMemPackage_InMemory(t *testing.T) {
	mp := &std.MemPackage{Name: "foo", Path: "gno.land/p/demo/foo"}
	p := packageFromMemPackage(mp)
	got, err := p.ToMemPackage()
	require.NoError(t, err)
	assert.Same(t, mp, got)
}

// A filesystem-backed package must re-read from disk on every ToMemPackage
// call: gnodev's hot reload depends on each genesis rebuild seeing the
// current on-disk content. A future memoization of the FS path would
// silently break reload and still pass the other tests, so pin the re-read.
func TestPackage_ToMemPackage_FS_RereadsOnEachCall(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnomod.toml"),
		fmt.Appendf(nil, "module = %q\n", "gno.land/p/demo/foo"), 0o644))
	foo := filepath.Join(dir, "foo.gno")
	require.NoError(t, os.WriteFile(foo, []byte("package foo\nfunc A() {}\n"), 0o644))

	p := &Package{ImportPath: "gno.land/p/demo/foo", Dir: dir, Kind: KindFS}

	first, err := p.ToMemPackage()
	require.NoError(t, err)
	require.NotNil(t, first.GetFile("foo.gno"))
	require.Equal(t, "package foo\nfunc A() {}\n", first.GetFile("foo.gno").Body)

	// Edit the file on disk; the next call must observe the new content.
	require.NoError(t, os.WriteFile(foo, []byte("package foo\nfunc B() {}\n"), 0o644))

	second, err := p.ToMemPackage()
	require.NoError(t, err)
	require.NotNil(t, second.GetFile("foo.gno"))
	assert.Equal(t, "package foo\nfunc B() {}\n", second.GetFile("foo.gno").Body,
		"ToMemPackage must re-read from disk so hot reload sees on-disk edits")
}
