package examplespkgfetcher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFetchPackage_FiletestRouting asserts the fetcher mirrors ReadMemPackage:
// any .gno file in filetests/ is loaded with its bare basename as MemFile.Name
// and Kind=KindFiletest (regardless of suffix); non-.gno files in filetests/
// are ignored.
func TestFetchPackage_FiletestRouting(t *testing.T) {
	t.Parallel()
	examplesDir := t.TempDir()
	pkgDir := filepath.Join(examplesDir, "gno.land", "p", "demo", "x")
	require.NoError(t, os.MkdirAll(filepath.Join(pkgDir, "filetests"), 0o755))

	write := func(rel, body string) {
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, rel), []byte(body), 0o644))
	}
	write("x.gno", "package x\n")
	// New-style: bare basename in filetests/.
	write(filepath.Join("filetests", "new.gno"), "package test\n")
	// Legacy-style: still works.
	write(filepath.Join("filetests", "legacy_filetest.gno"), "package test\n")
	// Non-.gno under filetests/ must be ignored.
	write(filepath.Join("filetests", "README.md"), "# nope\n")

	f := New(examplesDir)
	files, err := f.FetchPackage("gno.land/p/demo/x")
	require.NoError(t, err)

	byName := make(map[string]*std.MemFile, len(files))
	for _, m := range files {
		byName[m.Name] = m
	}
	assert.Contains(t, byName, "x.gno", "root .gno should be present")
	assert.Contains(t, byName, "new.gno", "filetest loaded with bare basename")
	assert.Contains(t, byName, "legacy_filetest.gno", "legacy-suffix filetest loaded")
	assert.NotContains(t, byName, "filetests/new.gno", "MemFile.Name must not carry the filetests/ prefix")
	assert.NotContains(t, byName, "filetests/README.md", "non-.gno under filetests/ must be ignored")
	assert.Equal(t, std.KindFiletest, byName["new.gno"].Kind, "Kind stamped from disk location")
	assert.Equal(t, std.KindFiletest, byName["legacy_filetest.gno"].Kind, "Kind stamped from disk location")
}
