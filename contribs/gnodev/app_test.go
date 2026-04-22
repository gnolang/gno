package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Test helpers

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// writeWorkspacePkg writes a gnomod.toml + gno file at dir. Used by
// loader-level integration tests.
func writeWorkspacePkg(t *testing.T, dir, module, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnomod.toml"),
		[]byte(fmt.Sprintf("module = %q\n", module)), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg.gno"),
		[]byte(body), 0o644))
}

// importPaths returns the ImportPath of each *packages.Package.
func importPaths(pkgs []*packages.Package) []string {
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		out[i] = p.ImportPath
	}
	return out
}

// ---- E2: workspace eager-load

func TestGnodev_Workspace_EagerLoad(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "gnowork.toml"), []byte(""), 0o644))
	writeWorkspacePkg(t, filepath.Join(root, "foo"), "gno.land/p/ws/foo", "package foo\n")
	t.Chdir(root)

	l := packages.New(packages.Config{
		Workspace: root,
		Logger:    discardLogger(),
	})
	pkgs, err := l.LoadWorkspace()
	require.NoError(t, err)
	assert.Contains(t, importPaths(pkgs), "gno.land/p/ws/foo")
}

// ---- E3: no-workspace / discovery mode

func TestGnodev_NoWorkspace_DiscoveryMode(t *testing.T) {
	extraRoot := t.TempDir()
	writeWorkspacePkg(t, filepath.Join(extraRoot, "extpkg"), "gno.land/p/ext/one", "package one\n")

	l := packages.New(packages.Config{
		Workspace:  "",
		Examples:   true,
		ExtraRoots: []string{extraRoot},
		Logger:     discardLogger(),
	})

	// LoadWorkspace returns nil,nil when no workspace is set.
	pkgs, err := l.LoadWorkspace()
	require.NoError(t, err)
	assert.Nil(t, pkgs)

	// Resolve against the extra root still succeeds.
	got, err := l.Resolve("gno.land/p/ext/one")
	require.NoError(t, err)
	assert.Equal(t, "gno.land/p/ext/one", got.ImportPath)
	assert.Equal(t, packages.KindFS, got.Kind)
}

// ---- E4: app-level fatal when no workspace + -no-examples + no -extra-root

func TestGnodev_NoWorkspace_NoExamples_ConfigError(t *testing.T) {
	// Move into a directory that is NOT inside any gno workspace.
	dir := t.TempDir()
	t.Chdir(dir)

	cfg := defaultLocalAppConfig
	cfg.noExamples = true
	// No extraRoots, no positional dirs. Chain is otherwise valid.
	cfg.deployKey = defaultDeployerAddress.String()

	app := NewApp(discardLogger(), &cfg, commands.NewTestIO())
	err := app.Setup(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nothing to load",
		"expected the fatal flag-combination error, got: %v", err)
}

// ---- E5: staging mode eager-loads workspace + extra roots via LoadAll

func TestGnodev_Staging_EagerAll(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "gnowork.toml"), []byte(""), 0o644))
	writeWorkspacePkg(t, filepath.Join(workspace, "w"), "gno.land/p/ws/one", "package one\n")

	extra := t.TempDir()
	writeWorkspacePkg(t, filepath.Join(extra, "e"), "gno.land/p/ext/two", "package two\n")

	// Examples: false to avoid depending on $GNOROOT/examples at test time.
	t.Chdir(workspace)
	l := packages.New(packages.Config{
		Workspace:  workspace,
		Examples:   false,
		ExtraRoots: []string{extra},
		Logger:     discardLogger(),
	})

	pkgs, err := l.LoadAll()
	require.NoError(t, err)
	paths := importPaths(pkgs)
	assert.Contains(t, paths, "gno.land/p/ws/one")
	assert.Contains(t, paths, "gno.land/p/ext/two")
}

// ---- E6: Reload preserves both workspace and proxy-resolved paths

func TestGnodev_Reload_AfterProxyHit(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "gnowork.toml"), []byte(""), 0o644))
	writeWorkspacePkg(t, filepath.Join(workspace, "w"), "gno.land/p/ws/only", "package only\n")

	extra := t.TempDir()
	writeWorkspacePkg(t, filepath.Join(extra, "q"), "gno.land/p/ext/proxy", "package proxy\n")

	t.Chdir(workspace)
	l := packages.New(packages.Config{
		Workspace:  workspace,
		ExtraRoots: []string{extra},
		Logger:     discardLogger(),
	})

	// Eager workspace load at startup.
	_, err := l.LoadWorkspace()
	require.NoError(t, err)

	// Simulate a proxy hit: Resolve a package outside the workspace.
	_, err = l.Resolve("gno.land/p/ext/proxy")
	require.NoError(t, err)

	// Reload should return both the workspace package and the tracked
	// proxy-resolved package.
	out, err := l.Reload()
	require.NoError(t, err)
	paths := importPaths(out)
	assert.Contains(t, paths, "gno.land/p/ws/only")
	assert.Contains(t, paths, "gno.land/p/ext/proxy")
}
