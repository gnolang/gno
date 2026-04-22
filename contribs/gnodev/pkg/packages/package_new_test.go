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

func TestNewPackage_ToMemPackage_FS(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnomod.toml"),
		[]byte(fmt.Sprintf("module = %q\n", "gno.land/p/demo/foo")), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "foo.gno"),
		[]byte("package foo\n"), 0o644))

	p := &NewPackage{ImportPath: "gno.land/p/demo/foo", Dir: dir, Kind: KindFS}
	mp, err := p.ToMemPackage()
	require.NoError(t, err)
	assert.Equal(t, "gno.land/p/demo/foo", mp.Path)
	assert.Equal(t, "foo", mp.Name)
}

func TestNewPackage_ToMemPackage_InMemory(t *testing.T) {
	mp := &std.MemPackage{Name: "foo", Path: "gno.land/p/demo/foo"}
	p := newPackageFromMemPackage(mp)
	got, err := p.ToMemPackage()
	require.NoError(t, err)
	assert.Same(t, mp, got)
}
