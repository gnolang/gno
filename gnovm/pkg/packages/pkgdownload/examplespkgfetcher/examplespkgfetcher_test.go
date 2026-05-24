package examplespkgfetcher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFetchPackage_FiletestsLayout asserts the fetcher mirrors ReadMemPackage:
// every .gno file in filetests/ is loaded with a "filetests/" prefix on its
// MemFile.Name, non-.gno files in filetests/ are ignored, and a foo.gno at the
// package root is not blocked by a co-named filetests/foo.gno.
func TestFetchPackage_FiletestsLayout(t *testing.T) {
	t.Parallel()
	examplesDir := t.TempDir()
	pkgDir := filepath.Join(examplesDir, "gno.land", "p", "demo", "x")
	require.NoError(t, os.MkdirAll(filepath.Join(pkgDir, "filetests"), 0o755))

	write := func(rel, body string) {
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, rel), []byte(body), 0o644))
	}
	write("x.gno", "package x\n")
	write(filepath.Join("filetests", "x.gno"), "package test\n")
	write(filepath.Join("filetests", "legacy_filetest.gno"), "package test\n")
	// Non-.gno under filetests/ must be ignored.
	write(filepath.Join("filetests", "README.md"), "# nope\n")

	f := New(examplesDir)
	files, err := f.FetchPackage("gno.land/p/demo/x")
	require.NoError(t, err)

	names := make(map[string]string, len(files))
	for _, m := range files {
		names[m.Name] = m.Body
	}
	assert.Contains(t, names, "x.gno", "root .gno should be present")
	assert.Contains(t, names, "filetests/x.gno", "filetests prefix must be encoded on Name")
	assert.Contains(t, names, "filetests/legacy_filetest.gno", "legacy suffix in filetests/ must still load")
	assert.NotContains(t, names, "filetests/README.md", "non-.gno under filetests/ must be ignored")
}
