package packages

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

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
	l := NewLoaderImpl(Config{Workspace: "", Logger: testLogger()})
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

	l := NewLoaderImpl(Config{Workspace: root, Logger: testLogger()})
	pkgs, err := l.LoadWorkspace()
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	assert.Equal(t, "gno.land/p/demo/foo", pkgs[0].ImportPath)
}
