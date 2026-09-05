package builder_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/misc/gnoe2e/internal/builder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalBuilder_BuildFromCheckout(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary; rerun without -short")
	}
	b := builder.NewLocalBuilder()
	outDir := t.TempDir()

	path, err := b.Build(context.Background(), builder.BuildOpts{OutDir: outDir})
	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.Equal(t, "gnoland", filepath.Base(path), "empty Binary should default to gnoland")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&0111, "binary should be executable")
}

// A build with no OutDir allocates its own temp directory, and the caller
// never learns its name. A failure that leaves it behind therefore leaves a
// directory nothing can ever find again, one per failed build.
func TestBuildWithoutOutDirLeavesNothingBehindOnFailure(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())

	_, err := builder.NewLocalBuilder().Build(context.Background(), builder.BuildOpts{Binary: "bogus"})
	require.Error(t, err)

	left, err := filepath.Glob(filepath.Join(os.TempDir(), "gnoe2e-build-*"))
	require.NoError(t, err)
	assert.Empty(t, left, "the build directory it created is its own to remove")
}

func TestBuildGpaoFromCheckout(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary; rerun without -short")
	}
	b := builder.NewLocalBuilder()

	path, err := b.Build(context.Background(), builder.BuildOpts{
		Binary: "gpao",
		OutDir: t.TempDir(),
	})
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&0o111, "gpao must be executable")
}

func TestBuildRejectsUnknownBinary(t *testing.T) {
	b := builder.NewLocalBuilder()

	_, err := b.Build(context.Background(), builder.BuildOpts{
		Binary: "gnoweb",
		OutDir: t.TempDir(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "gpao")
}
