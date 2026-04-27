package packages

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
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

// TestLoader_LoadAll_LogsProgress verifies LoadAll emits progress per root
// so the user knows it isn't hung.
func TestLoader_LoadAll_LogsProgress(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))
	writePkg(t, filepath.Join(root, "p"), "gno.land/p/ws/one", "package one\n")

	extra := t.TempDir()
	writePkg(t, filepath.Join(extra, "q"), "gno.land/p/ext/two", "package two\n")

	t.Chdir(root)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	l := New(Config{Workspace: root, ExtraRoots: []string{extra}, Logger: logger})
	_, err := l.LoadAll()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "loading root", "should log per-root progress")
	assert.Contains(t, out, extra, "progress log should name the root")
	// Progress is emitted as structured kv (n=<i>, of=<total>), not a
	// formatted "1/1" string, so each field can be filtered independently.
	assert.Contains(t, out, "n=1", "progress should expose n as a structured field")
	assert.Contains(t, out, "of=1", "progress should expose total as a structured field")
}

// TestPackage_KindZeroValue verifies the zero value of Kind is KindUnknown,
// not KindFS. This makes a forgotten Kind field trip loudly rather than
// silently registering the package as filesystem-backed.
func TestPackage_KindZeroValue(t *testing.T) {
	var p Package
	assert.Equal(t, KindUnknown, p.Kind)
	assert.NotEqual(t, KindFS, p.Kind)
}

// TestLoader_LoadRealExamplesRealm exercises loading a real realm from
// $GNOROOT/examples. boards2/v1 imports chain, chain/runtime, p-tree, etc.
// — the kind of graph that triggers stripStdlibs + MPUserProd code paths
// that trivial single-package tests miss. Skips cleanly if the realm path
// doesn't exist (e.g., running outside the monorepo).
func TestLoader_LoadRealExamplesRealm(t *testing.T) {
	gnoroot := os.Getenv("GNOROOT")
	if gnoroot == "" {
		// Fall back to gnoenv discovery. Test target is a stable example.
		gnoroot = filepath.Join("..", "..", "..", "..")
	}
	realmDir := filepath.Join(gnoroot, "examples", "gno.land", "r", "gnoland", "boards2", "v1")
	if _, err := os.Stat(realmDir); err != nil {
		t.Skipf("examples realm not available: %v", err)
	}
	absRealm, err := filepath.Abs(realmDir)
	require.NoError(t, err)

	// Set up a workspace at the realm dir (boards2/v1 has its own gnomod.toml).
	t.Chdir(absRealm)

	l := New(Config{
		Workspace: absRealm,
		Examples:  true,
		GnoRoot:   filepath.Join(absRealm, "..", "..", "..", "..", ".."),
		Logger:    testLogger(),
	})
	pkgs, err := l.LoadWorkspace()
	require.NoError(t, err, "boards2/v1 should load without errors")
	require.NotEmpty(t, pkgs, "should resolve at least one package")

	// Verify it loaded the realm itself.
	paths := pathsOf(pkgs)
	assert.Contains(t, paths, "gno.land/r/gnoland/boards2/v1")

	// Verify ToMemPackage works for the realm (regression for MPUserProd fix).
	for _, p := range pkgs {
		if p.ImportPath == "gno.land/r/gnoland/boards2/v1" {
			mp, err := p.ToMemPackage()
			require.NoError(t, err, "ToMemPackage must succeed on real realm")
			assert.NotEmpty(t, mp.Name, "MemPackage must have a Name")
			// MPUserProd → no _test.gno files. Regression for the bug fixed
			// in commit 62e4e3246f.
			for _, f := range mp.Files {
				assert.NotContains(t, f.Name, "_test.gno",
					"MPUserProd must strip test files; got %s", f.Name)
			}
			return
		}
	}
	t.Fatalf("boards2/v1 not found in loaded packages: %v", paths)
}

// recordingFetcher counts FetchPackage invocations so tests can assert that
// LookupFS — which is FS-only — never reaches the rpc fetcher.
type recordingFetcher struct {
	calls atomic.Int32
}

func (f *recordingFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	f.calls.Add(1)
	return nil, fmt.Errorf("not in test fetcher: %s", pkgPath)
}

// TestLoader_LookupFS_NoFetcherCall locks in the contract that LookupFS is
// FS-only: the rpc fetcher must never be invoked, even on a miss. Resolve
// would have called it on miss; LookupFS must not.
func TestLoader_LookupFS_NoFetcherCall(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "alone")
	writePkg(t, pkgDir, "gno.land/p/me/alone", "package alone\n")

	rec := &recordingFetcher{}
	l := New(Config{
		ExtraRoots: []string{root},
		Fetcher:    rec,
		Logger:     testLogger(),
	})

	// Hit
	assert.True(t, l.LookupFS("gno.land/p/me/alone"))
	// Miss (would have triggered fetcher in Resolve)
	assert.False(t, l.LookupFS("gno.land/r/never/exists"))

	assert.Zero(t, rec.calls.Load(), "LookupFS must never invoke the fetcher")
}
