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

func TestLocalBuilder_UnknownBinaryError(t *testing.T) {
	b := builder.NewLocalBuilder()

	ctx := context.Background()
	_, err := b.Build(ctx, builder.BuildOpts{Binary: "bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown binary")
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
