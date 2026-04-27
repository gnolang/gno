package packages

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func writePkg(t *testing.T, dir, module, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	modToml := fmt.Sprintf("module = %q\n", module)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnomod.toml"),
		[]byte(modToml), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg.gno"),
		[]byte(body), 0o644))
}

func TestLoader_LoadWorkspace_Empty(t *testing.T) {
	l := New(Config{Workspace: "", Logger: testLogger()})
	pkgs, err := l.LoadWorkspace()
	require.NoError(t, err)
	assert.Empty(t, pkgs)
}

func TestLoader_LoadWorkspace_OnePackage(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "demo")
	writePkg(t, pkgDir, "gno.land/p/demo/foo", "package foo\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))

	t.Chdir(root)

	l := New(Config{Workspace: root, Logger: testLogger()})
	pkgs, err := l.LoadWorkspace()
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	assert.Equal(t, "gno.land/p/demo/foo", pkgs[0].ImportPath)
}

// TestLoader_LoadWorkspace_WithStdlibImport exercises the stripStdlibs path:
// a workspace package importing a native stdlib (like "chain") must not
// cause PkgList.Sort to fail on the missing dep (gnovm.Load skips native
// stdlibs during dep traversal but leaves them in each pkg's Imports).
func TestLoader_LoadWorkspace_WithStdlibImport(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "demo")
	writePkg(t, pkgDir, "gno.land/p/demo/bar",
		`package bar
import "chain"
var _ = chain.ChainDomain
`)
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))

	t.Chdir(root)

	l := New(Config{Workspace: root, Logger: testLogger()})
	pkgs, err := l.LoadWorkspace()
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	assert.Equal(t, "gno.land/p/demo/bar", pkgs[0].ImportPath)
}

func TestLoader_Resolve_IndexHit(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "demo")
	writePkg(t, pkgDir, "gno.land/p/demo/foo", "package foo\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))

	t.Chdir(root)

	l := New(Config{Workspace: root, Logger: testLogger()})
	_, err := l.LoadWorkspace()
	require.NoError(t, err)

	got, err := l.Resolve("gno.land/p/demo/foo")
	require.NoError(t, err)
	assert.Equal(t, "gno.land/p/demo/foo", got.ImportPath)
	assert.Equal(t, pkgDir, got.Dir)
}

func TestLoader_Resolve_MissReturnsNotFound(t *testing.T) {
	// Empty fetcher so the RPC fallback fails fast (no real network calls).
	l := New(Config{
		Fetcher: pkgdownload.NewInMemoryFetcher(),
		Logger:  testLogger(),
	})
	_, err := l.Resolve("gno.land/p/absent")
	assert.ErrorIs(t, err, ErrPackageNotFound)
}

func TestLoader_Resolve_FSWalk(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "mypkg")
	writePkg(t, pkgDir, "gno.land/p/custom/mypkg", "package mypkg\n")

	l := New(Config{ExtraRoots: []string{root}, Logger: testLogger()})
	got, err := l.Resolve("gno.land/p/custom/mypkg")
	require.NoError(t, err)
	assert.Equal(t, pkgDir, got.Dir)

	// second call hits the index
	got2, err := l.Resolve("gno.land/p/custom/mypkg")
	require.NoError(t, err)
	assert.Same(t, got, got2)
}

func TestLoader_Resolve_RPCFallback(t *testing.T) {
	mp := &std.MemPackage{
		Path:  "gno.land/r/demo/boards",
		Name:  "boards",
		Files: []*std.MemFile{{Name: "boards.gno", Body: "package boards\n"}},
	}
	l := New(Config{
		Fetcher: pkgdownload.NewInMemoryFetcher(mp),
		Logger:  testLogger(),
	})

	got, err := l.Resolve("gno.land/r/demo/boards")
	require.NoError(t, err)
	assert.Equal(t, KindRemote, got.Kind)
	assert.Equal(t, "gno.land/r/demo/boards", got.ImportPath)
}

func TestLoader_Reload_IncludesTrackedPaths(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))
	wsPkg := filepath.Join(root, "wspkg")
	writePkg(t, wsPkg, "gno.land/p/ws/one", "package one\n")

	extra := t.TempDir()
	extraPkg := filepath.Join(extra, "p")
	writePkg(t, extraPkg, "gno.land/p/ext/two", "package two\n")

	t.Chdir(root)

	l := New(Config{Workspace: root, ExtraRoots: []string{extra}, Logger: testLogger()})
	_, err := l.LoadWorkspace()
	require.NoError(t, err)
	_, err = l.Resolve("gno.land/p/ext/two")
	require.NoError(t, err)

	got, err := l.Reload()
	require.NoError(t, err)
	paths := pathsOf(got)
	assert.Contains(t, paths, "gno.land/p/ws/one")
	assert.Contains(t, paths, "gno.land/p/ext/two")
}

func pathsOf(pkgs []*Package) []string {
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		out[i] = p.ImportPath
	}
	return out
}

func TestLoader_LoadAll(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))
	writePkg(t, filepath.Join(root, "p"), "gno.land/p/ws/one", "package one\n")

	extra := t.TempDir()
	writePkg(t, filepath.Join(extra, "q"), "gno.land/p/ext/two", "package two\n")

	t.Chdir(root)

	l := New(Config{Workspace: root, ExtraRoots: []string{extra}, Logger: testLogger()})
	pkgs, err := l.LoadAll()
	require.NoError(t, err)
	paths := pathsOf(pkgs)
	assert.Contains(t, paths, "gno.land/p/ws/one")
	assert.Contains(t, paths, "gno.land/p/ext/two")
}

// TestLoader_Reload_PreservesRootIdx verifies that Reload does NOT invalidate
// rootIdx — directories are stable mid-session and re-walking large roots on
// every watcher event is too expensive. Restart is required to pick up new dirs.
func TestLoader_Reload_PreservesRootIdx(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))
	wsPkg := filepath.Join(root, "wspkg")
	writePkg(t, wsPkg, "gno.land/p/ws/one", "package one\n")

	extra := t.TempDir()
	extraPkg := filepath.Join(extra, "p")
	writePkg(t, extraPkg, "gno.land/p/ext/two", "package two\n")

	t.Chdir(root)

	l := New(Config{Workspace: root, ExtraRoots: []string{extra}, Logger: testLogger()})
	_, err := l.LoadWorkspace()
	require.NoError(t, err)
	_, err = l.Resolve("gno.land/p/ext/two")
	require.NoError(t, err)

	// Snapshot rootIdx before Reload.
	l.mu.RLock()
	idxBefore, ok := l.rootIdx[extra]
	require.True(t, ok, "rootIdx should contain the extra root after Resolve")
	require.NotEmpty(t, idxBefore)
	l.mu.RUnlock()

	_, err = l.Reload()
	require.NoError(t, err)

	// rootIdx must still be populated after Reload.
	l.mu.RLock()
	idxAfter, ok := l.rootIdx[extra]
	l.mu.RUnlock()
	// rootIdx must still be the SAME map across Reload, not just an
	// equivalent freshly re-walked one. Compare map header pointers.
	require.True(t, ok, "rootIdx should be preserved across Reload")
	assert.Equal(t,
		reflect.ValueOf(idxBefore).Pointer(),
		reflect.ValueOf(idxAfter).Pointer(),
		"rootIdx must be the same map (no re-walk), not a content-equivalent rebuild",
	)
}
