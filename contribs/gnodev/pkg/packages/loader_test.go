package packages

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync/atomic"
	"testing"

	vmpackages "github.com/gnolang/gno/gnovm/pkg/packages"
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

// TestLoader_Reload_SingleModuleWorkspace covers the canonical
// `cd myrealm && gnodev` flow: a directory with gnomod.toml but no
// gnowork.toml ancestor. gnovm treats that as single-package mode and
// rejects recursive patterns, so the loader must not blindly append "/...".
func TestLoader_Reload_SingleModuleWorkspace(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "ossas")
	writePkg(t, dir, "gno.land/r/ossas", "package ossas\n")

	t.Chdir(dir)

	ws := FindWorkspace(dir)
	require.Equal(t, dir, ws, "gnomod.toml dir must be detected as workspace root")

	l := New(Config{Workspace: ws, Logger: testLogger()})
	pkgs, err := l.Reload()
	require.NoError(t, err)
	assert.Equal(t, []string{"gno.land/r/ossas"}, pathsOf(pkgs))
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

// TestLoader_Reload_TrackedPathIncludesFSDeps reproduces the lazy-load flow:
// the proxy Resolves a single realm, then Reload builds the genesis set.
// The realm's transitive imports must be included (dependencies first), or
// its genesis addpkg fails type-checking with "unknown import path".
func TestLoader_Reload_TrackedPathIncludesFSDeps(t *testing.T) {
	gnoroot := t.TempDir()
	examples := filepath.Join(gnoroot, "examples")
	writePkg(t, filepath.Join(examples, "base"), "gno.land/p/test/base",
		"package base\nfunc Hi() string { return \"hi\" }\n")
	writePkg(t, filepath.Join(examples, "dep"), "gno.land/p/test/dep",
		`package dep
import (
	"strings"
	"gno.land/p/test/base"
)
func Hello() string { return strings.ToUpper(base.Hi()) }
`)
	writePkg(t, filepath.Join(examples, "home"), "gno.land/r/test/home",
		`package home
import "gno.land/p/test/dep"
func Render(_ string) string { return dep.Hello() }
`)

	l := New(Config{Examples: true, GnoRoot: gnoroot, Logger: testLogger()})
	_, err := l.Resolve("gno.land/r/test/home")
	require.NoError(t, err)

	got, err := l.Reload()
	require.NoError(t, err)

	paths := pathsOf(got)
	require.ElementsMatch(t, paths,
		[]string{"gno.land/p/test/base", "gno.land/p/test/dep", "gno.land/r/test/home"})
	assert.Less(t, slices.Index(paths, "gno.land/p/test/base"), slices.Index(paths, "gno.land/p/test/dep"))
	assert.Less(t, slices.Index(paths, "gno.land/p/test/dep"), slices.Index(paths, "gno.land/r/test/home"))
}

func TestLoader_Reload_TrackedPathIncludesRemoteDeps(t *testing.T) {
	dep := &std.MemPackage{
		Path:  "gno.land/p/test/dep",
		Name:  "dep",
		Files: []*std.MemFile{{Name: "dep.gno", Body: "package dep\nfunc Hello() string { return \"hi\" }\n"}},
	}
	home := &std.MemPackage{
		Path: "gno.land/r/test/home",
		Name: "home",
		Files: []*std.MemFile{{Name: "home.gno", Body: `package home
import "gno.land/p/test/dep"
func Render(_ string) string { return dep.Hello() }
`}},
	}

	l := New(Config{
		Fetcher: pkgdownload.NewInMemoryFetcher(dep, home),
		Logger:  testLogger(),
	})
	_, err := l.Resolve("gno.land/r/test/home")
	require.NoError(t, err)

	got, err := l.Reload()
	require.NoError(t, err)

	paths := pathsOf(got)
	require.ElementsMatch(t, paths, []string{"gno.land/p/test/dep", "gno.land/r/test/home"})
	assert.Less(t, slices.Index(paths, "gno.land/p/test/dep"), slices.Index(paths, "gno.land/r/test/home"))
}

// TestLoader_Reload_TrackedPathMissingDep: an unresolvable import must not
// abort the reload; the package is still returned and the chain reports the
// precise type-check error at deploy time.
func TestLoader_Reload_TrackedPathMissingDep(t *testing.T) {
	gnoroot := t.TempDir()
	examples := filepath.Join(gnoroot, "examples")
	writePkg(t, filepath.Join(examples, "home"), "gno.land/r/test/home",
		`package home
import "gno.land/p/test/absent"
func Render(_ string) string { return absent.Hello() }
`)

	l := New(Config{
		Examples: true,
		GnoRoot:  gnoroot,
		Fetcher:  pkgdownload.NewInMemoryFetcher(),
		Logger:   testLogger(),
	})
	_, err := l.Resolve("gno.land/r/test/home")
	require.NoError(t, err)

	got, err := l.Reload()
	require.NoError(t, err)
	assert.Equal(t, []string{"gno.land/r/test/home"}, pathsOf(got))
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

// TestLoader_LoadAll_SortsExtraRoots verifies LoadAll returns packages in
// topological order across extra roots (deps before dependents). Genesis
// deploy applies packages in slice order, so any dependent appearing before
// its dep fails type-checking.
func TestLoader_LoadAll_SortsExtraRoots(t *testing.T) {
	extra := t.TempDir()
	// Chain: aa imports bb imports cc imports dd imports ee. Package names
	// must match ^[a-z][a-z0-9_]+$ (validatePkgName, nodes.go), so two
	// chars minimum. 5 entries give 120 permutations of map iteration
	// order, only one of which matches topological order — wide enough to
	// catch a missing sort step.
	writePkg(t, filepath.Join(extra, "aa"), "gno.land/p/ext/aa",
		"package aa\nimport _ \"gno.land/p/ext/bb\"\n")
	writePkg(t, filepath.Join(extra, "bb"), "gno.land/p/ext/bb",
		"package bb\nimport _ \"gno.land/p/ext/cc\"\n")
	writePkg(t, filepath.Join(extra, "cc"), "gno.land/p/ext/cc",
		"package cc\nimport _ \"gno.land/p/ext/dd\"\n")
	writePkg(t, filepath.Join(extra, "dd"), "gno.land/p/ext/dd",
		"package dd\nimport _ \"gno.land/p/ext/ee\"\n")
	writePkg(t, filepath.Join(extra, "ee"), "gno.land/p/ext/ee",
		"package ee\n")

	// LoadAll requires a workspace context for findLoaderContext. Use an
	// empty workspace dir so loadWithPatterns is a no-op and we exercise
	// only the extra-root sort path.
	ws := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(ws, "gnowork.toml"), []byte(""), 0o644))
	t.Chdir(ws)

	l := New(Config{Workspace: ws, ExtraRoots: []string{extra}, Logger: testLogger()})
	pkgs, err := l.LoadAll()
	require.NoError(t, err)

	pos := map[string]int{}
	for i, p := range pkgs {
		pos[p.ImportPath] = i
	}
	for _, chain := range [][2]string{
		{"gno.land/p/ext/ee", "gno.land/p/ext/dd"},
		{"gno.land/p/ext/dd", "gno.land/p/ext/cc"},
		{"gno.land/p/ext/cc", "gno.land/p/ext/bb"},
		{"gno.land/p/ext/bb", "gno.land/p/ext/aa"},
	} {
		dep, dependent := chain[0], chain[1]
		di, ok := pos[dep]
		require.True(t, ok, "%s missing from LoadAll output", dep)
		dni, ok := pos[dependent]
		require.True(t, ok, "%s missing from LoadAll output", dependent)
		assert.Less(t, di, dni, "%s (dep) must come before %s (dependent); got order %v",
			dep, dependent, pathsOf(pkgs))
	}
}

// TestLoader_LoadAll_SkipsIgnoredExtraRootPkgs verifies that a package in an
// extra root whose gnomod.toml sets `ignore = true` is filtered out by
// GetNonIgnoredPkgs. The Ignore flag must reach the synthesized
// vmpackages.Package so the sort+filter chain in LoadAll drops it before
// genesis deploy.
func TestLoader_LoadAll_SkipsIgnoredExtraRootPkgs(t *testing.T) {
	extra := t.TempDir()
	writePkg(t, filepath.Join(extra, "live"), "gno.land/p/ext/live",
		"package live\n")

	// Ignored package: writePkg only emits `module = ...`; append the ignore line.
	dropDir := filepath.Join(extra, "dropme")
	writePkg(t, dropDir, "gno.land/p/ext/dropme", "package dropme\n")
	mod, err := os.ReadFile(filepath.Join(dropDir, "gnomod.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dropDir, "gnomod.toml"),
		append(mod, []byte("ignore = true\n")...), 0o644))

	ws := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(ws, "gnowork.toml"), []byte(""), 0o644))
	t.Chdir(ws)

	l := New(Config{Workspace: ws, ExtraRoots: []string{extra}, Logger: testLogger()})
	pkgs, err := l.LoadAll()
	require.NoError(t, err)
	paths := pathsOf(pkgs)
	assert.Contains(t, paths, "gno.land/p/ext/live")
	assert.NotContains(t, paths, "gno.land/p/ext/dropme",
		"ignore=true extra-root pkg must be filtered out before deploy")
}

// TestLoader_Reload_EagerLoadsExtraRootDeps verifies that Reload eagerly
// materializes packages in -extra-root directories so cross-package
// dependencies within an extra-root resolve at startup, not only via the
// lazy proxy. Without this, a realm in an extra-root that imports a
// sibling pure-package in the same extra-root fails to compile on first
// query because the dep was never deployed to the chain.
func TestLoader_Reload_EagerLoadsExtraRootDeps(t *testing.T) {
	extra := t.TempDir()
	writePkg(t, filepath.Join(extra, "lib"), "gno.land/p/eager/lib",
		"package lib\n\nfunc Greet() string { return \"hi\" }\n")
	writePkg(t, filepath.Join(extra, "realm"), "gno.land/r/eager/realm",
		"package realm\n\nimport \"gno.land/p/eager/lib\"\n\nfunc Render(_ string) string { return lib.Greet() }\n")

	ws := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(ws, "gnowork.toml"), []byte(""), 0o644))
	t.Chdir(ws)

	l := New(Config{
		Workspace:  ws,
		ExtraRoots: []string{extra},
		Fetcher:    pkgdownload.NewInMemoryFetcher(),
		Logger:     testLogger(),
	})

	pkgs, err := l.Reload()
	require.NoError(t, err)
	paths := pathsOf(pkgs)
	assert.Contains(t, paths, "gno.land/p/eager/lib",
		"extra-root pure-package must be eagerly loaded by Reload")
	assert.Contains(t, paths, "gno.land/r/eager/realm",
		"extra-root realm must be eagerly loaded by Reload")

	pos := map[string]int{}
	for i, p := range pkgs {
		pos[p.ImportPath] = i
	}
	assert.Less(t, pos["gno.land/p/eager/lib"], pos["gno.land/r/eager/realm"],
		"dep (lib) must sort before its dependent (realm); got %v", paths)
}

// TestLoader_ExcludeDirs_SkipsSubtree verifies Config.ExcludeDirs causes
// scanRoot to skip the named directories: packages under an excluded path
// must not appear via Resolve (FS walk) or LoadAll (eager root traversal).
// gnodev uses this to honor -without-quarantined-examples by passing
// $GNOROOT/examples/quarantined as an excluded dir.
func TestLoader_ExcludeDirs_SkipsSubtree(t *testing.T) {
	extra := t.TempDir()
	writePkg(t, filepath.Join(extra, "live", "foo"), "gno.land/p/live/foo",
		"package foo\n")
	skipDir := filepath.Join(extra, "skipme")
	writePkg(t, filepath.Join(skipDir, "bar"), "gno.land/p/skipme/bar",
		"package bar\n")

	ws := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(ws, "gnowork.toml"), []byte(""), 0o644))
	t.Chdir(ws)

	baseCfg := Config{
		Workspace:  ws,
		ExtraRoots: []string{extra},
		Fetcher:    pkgdownload.NewInMemoryFetcher(),
		Logger:     testLogger(),
	}

	// Control: without ExcludeDirs the skipme/bar package resolves via FS.
	// Establishes that ErrPackageNotFound below is caused by the exclude,
	// not by an unrelated fetcher/index gap.
	ctrl := New(baseCfg)
	ctrlBar, err := ctrl.Resolve("gno.land/p/skipme/bar")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(skipDir, "bar"), ctrlBar.Dir)

	excludedCfg := baseCfg
	excludedCfg.ExcludeDirs = []string{skipDir}
	l := New(excludedCfg)

	got, err := l.Resolve("gno.land/p/live/foo")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(extra, "live", "foo"), got.Dir)

	_, err = l.Resolve("gno.land/p/skipme/bar")
	assert.ErrorIs(t, err, ErrPackageNotFound)

	pkgs, err := l.LoadAll()
	require.NoError(t, err)
	paths := pathsOf(pkgs)
	assert.Contains(t, paths, "gno.land/p/live/foo")
	assert.NotContains(t, paths, "gno.land/p/skipme/bar",
		"ExcludeDirs must skip the subtree during LoadAll's root scan")
}

// TestLoader_LoadAll_LogsProgress verifies LoadAll emits per-root progress
// events; users opt into seeing them with -v (Debug level).
func TestLoader_LoadAll_LogsProgress(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))
	writePkg(t, filepath.Join(root, "p"), "gno.land/p/ws/one", "package one\n")

	extra := t.TempDir()
	writePkg(t, filepath.Join(extra, "q"), "gno.land/p/ext/two", "package two\n")

	t.Chdir(root)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

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

	// ToMemPackage uses MPUserProd, which strips _test.gno files. Realms
	// like boards2/v1 have test files that import not-yet-deployed packages;
	// shipping them would fail chain-side type checks at deploy time.
	for _, p := range pkgs {
		if p.ImportPath == "gno.land/r/gnoland/boards2/v1" {
			mp, err := p.ToMemPackage()
			require.NoError(t, err, "ToMemPackage must succeed on real realm")
			assert.NotEmpty(t, mp.Name, "MemPackage must have a Name")
			for _, f := range mp.Files {
				assert.NotContains(t, f.Name, "_test.gno",
					"MPUserProd must strip test files; got %s", f.Name)
			}
			return
		}
	}
	t.Fatalf("boards2/v1 not found in loaded packages: %v", paths)
}

// TestLoader_Reload_ExamplesDepsFromDisk: a workspace package importing an
// examples-resident dep must get it from $GNOROOT/examples, never the
// fetcher — on-disk source is fresher than the chain and works offline.
// The fetcher here errors on any call, so a network attempt fails loudly.
func TestLoader_Reload_ExamplesDepsFromDisk(t *testing.T) {
	gnoroot := t.TempDir()
	writePkg(t, filepath.Join(gnoroot, "examples", "lib"), "gno.land/p/test/lib",
		"package lib\nfunc Hi() string { return \"hi\" }\n")

	ws := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(ws, "gnowork.toml"), []byte(""), 0o644))
	writePkg(t, filepath.Join(ws, "realm"), "gno.land/r/test/realm",
		`package realm
import "gno.land/p/test/lib"
func Render(_ string) string { return lib.Hi() }
`)
	t.Chdir(ws)

	rec := &recordingFetcher{}
	l := New(Config{
		Workspace: ws,
		Examples:  true,
		GnoRoot:   gnoroot,
		Fetcher:   rec,
		Logger:    testLogger(),
	})

	pkgs, err := l.Reload()
	require.NoError(t, err)
	paths := pathsOf(pkgs)
	assert.Contains(t, paths, "gno.land/r/test/realm")
	assert.Contains(t, paths, "gno.land/p/test/lib",
		"examples-resident dep must be resolved from disk")
	assert.Zero(t, rec.calls.Load(),
		"dep sits in $GNOROOT/examples; the fetcher must not be consulted")
}

// TestLoader_AddLocalPackage covers the gnomod-less dir flow (`gnodev
// ./scratch-realm` or `cd scratch-realm && gnodev`): the dir is registered
// under a generated module path, reaches every reload, and its MemPackage
// carries a synthesized gnomod.toml — chain-side AddPackage validation
// rejects packages without one.
func TestLoader_AddLocalPackage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myrealm")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "realm.gno"),
		[]byte("package myrealm\n"), 0o644))

	l := New(Config{Fetcher: pkgdownload.NewInMemoryFetcher(), Logger: testLogger()})
	l.AddLocalPackage("gno.land/r/dev/myrealm", dir)

	got, err := l.Resolve("gno.land/r/dev/myrealm")
	require.NoError(t, err)
	assert.Equal(t, dir, got.Dir)

	pkgs, err := l.Reload()
	require.NoError(t, err)
	assert.Contains(t, pathsOf(pkgs), "gno.land/r/dev/myrealm",
		"registered local package must reach the reload output")

	mp, err := got.ToMemPackage()
	require.NoError(t, err)
	gm := mp.GetFile("gnomod.toml")
	require.NotNil(t, gm, "gnomod.toml must be synthesized for deploy")
	assert.Contains(t, gm.Body, `module = "gno.land/r/dev/myrealm"`)
	assert.NoError(t, mp.ValidateBasic(),
		"synthesized file must keep the mempackage valid (sorted files)")
}

// TestLoader_Track_ReloadIncludesTracked: paths registered via Track (the
// -paths flag and -txs-file dependencies) must be part of every Reload
// output, exactly like proxy-resolved paths.
func TestLoader_Track_ReloadIncludesTracked(t *testing.T) {
	gnoroot := t.TempDir()
	examples := filepath.Join(gnoroot, "examples")
	writePkg(t, filepath.Join(examples, "lib"), "gno.land/p/test/lib",
		"package lib\nfunc Hi() string { return \"hi\" }\n")
	writePkg(t, filepath.Join(examples, "realm"), "gno.land/r/test/realm",
		`package realm
import "gno.land/p/test/lib"
func Render(_ string) string { return lib.Hi() }
`)

	l := New(Config{Examples: true, GnoRoot: gnoroot, Logger: testLogger()})
	l.Track("gno.land/r/test/realm")

	got, err := l.Reload()
	require.NoError(t, err)
	assert.ElementsMatch(t, pathsOf(got),
		[]string{"gno.land/p/test/lib", "gno.land/r/test/realm"},
		"tracked path and its transitive deps must reach the reload output")
}

// TestLoader_Track_LoadAllIncludesTracked: staging mode reloads via LoadAll;
// tracked paths resolvable only through the fetcher must still be included.
func TestLoader_Track_LoadAllIncludesTracked(t *testing.T) {
	mp := &std.MemPackage{
		Path:  "gno.land/r/remote/only",
		Name:  "only",
		Files: []*std.MemFile{{Name: "only.gno", Body: "package only\n"}},
	}

	ws := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(ws, "gnowork.toml"), []byte(""), 0o644))
	t.Chdir(ws)

	l := New(Config{
		Workspace: ws,
		Fetcher:   pkgdownload.NewInMemoryFetcher(mp),
		Logger:    testLogger(),
	})
	l.Track("gno.land/r/remote/only")

	got, err := l.LoadAll()
	require.NoError(t, err)
	assert.Contains(t, pathsOf(got), "gno.land/r/remote/only",
		"LoadAll must include tracked paths")
}

// TestLoader_Reload_KeepsRemoteCached: on-chain packages are immutable for a
// gnodev session, and Reload runs on every watcher tick. Evicting remote
// entries from the index would re-fetch them over RPC on every file save.
func TestLoader_Reload_KeepsRemoteCached(t *testing.T) {
	mp := &std.MemPackage{
		Path:  "gno.land/r/demo/boards",
		Name:  "boards",
		Files: []*std.MemFile{{Name: "boards.gno", Body: "package boards\n"}},
	}
	cf := &recordingFetcher{inner: pkgdownload.NewInMemoryFetcher(mp)}
	l := New(Config{Fetcher: cf, Logger: testLogger()})

	_, err := l.Resolve("gno.land/r/demo/boards")
	require.NoError(t, err)
	require.EqualValues(t, 1, cf.calls.Load())

	for range 2 {
		got, err := l.Reload()
		require.NoError(t, err)
		assert.Contains(t, pathsOf(got), "gno.land/r/demo/boards",
			"tracked remote package must stay in the reload output")
	}
	assert.EqualValues(t, 1, cf.calls.Load(),
		"remote packages are session-immutable; Reload must not re-fetch them")
}

// TestStripStdlibs_FiltersImportsSpecs verifies stripStdlibs keeps Imports
// and ImportsSpecs consistent: GetNonIgnoredPkgs walks ImportsSpecs, so a
// stdlib entry surviving there while gone from Imports is a desync trap.
func TestStripStdlibs_FiltersImportsSpecs(t *testing.T) {
	pkg := &vmpackages.Package{
		ImportPath: "gno.land/p/test/pkg",
		Imports: map[vmpackages.FileKind][]string{
			vmpackages.FileKindPackageSource: {"chain", "gno.land/p/test/dep"},
		},
		ImportsSpecs: vmpackages.ImportsMap{
			vmpackages.FileKindPackageSource: {
				{PkgPath: "chain"},
				{PkgPath: "gno.land/p/test/dep"},
			},
		},
	}

	out := stripStdlibs(vmpackages.PkgList{pkg})

	require.Len(t, out, 1)
	assert.Equal(t, []string{"gno.land/p/test/dep"},
		out[0].Imports[vmpackages.FileKindPackageSource])
	assert.Equal(t, []string{"gno.land/p/test/dep"},
		out[0].ImportsSpecs.ToStrings()[vmpackages.FileKindPackageSource],
		"ImportsSpecs must drop stdlib entries alongside Imports")
}

// TestLoader_KindForDir_ModCacheBoundary verifies the modcache check matches
// on path-segment boundaries: a sibling directory whose name merely starts
// with the modcache path must classify as FS, not Remote.
func TestLoader_KindForDir_ModCacheBoundary(t *testing.T) {
	l := New(Config{Logger: testLogger()})
	l.modCache = filepath.Join("/x", "gnomodcache")
	l.modCachePrefix = l.modCache + string(filepath.Separator)

	assert.Equal(t, KindRemote, l.kindForDir(filepath.Join("/x", "gnomodcache", "gno.land", "p", "foo")))
	assert.Equal(t, KindRemote, l.kindForDir(filepath.Join("/x", "gnomodcache")))
	assert.Equal(t, KindFS, l.kindForDir(filepath.Join("/x", "gnomodcache-other", "pkg")))
}

// recordingFetcher counts FetchPackage invocations, delegating to inner when
// set and erroring otherwise. Lets tests assert which calls reach the rpc
// fetcher (LookupFS never may; Reload must not re-fetch cached remotes).
type recordingFetcher struct {
	inner pkgdownload.PackageFetcher
	calls atomic.Int32
}

func (f *recordingFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	f.calls.Add(1)
	if f.inner != nil {
		return f.inner.FetchPackage(pkgPath)
	}
	return nil, fmt.Errorf("not in test fetcher: %s", pkgPath)
}

// TestLoader_LookupFS_NoFetcherCall asserts the contract that LookupFS is
// FS-only: the rpc fetcher must never be invoked on a hit nor on a miss.
// recordingFetcher's call counter stays at zero across both paths.
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

	// Hit path.
	assert.True(t, l.LookupFS("gno.land/p/me/alone"))
	// Miss path; the FS-only contract forbids any fallback to the fetcher.
	assert.False(t, l.LookupFS("gno.land/r/never/exists"))

	assert.Zero(t, rec.calls.Load(), "LookupFS must never invoke the fetcher")
}
