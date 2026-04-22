package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
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
