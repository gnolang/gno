package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/require"
)

// lint must act on a package before it is read, so it takes its file list from
// MemPackageFilePaths -- the same function ReadMemPackage uses. This pins the
// extraction: the list and the package it produces must not disagree. While
// lint re-derived the list instead, it drifted twice, missing filetests/ and
// then over-including plain .gno files there.
func TestMemPackageFilePathsMatchesReadMemPackage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "filetests"), 0o755))
	write := func(rel, body string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644))
	}
	write("gnomod.toml", "module = \"gno.land/p/demo/zz\"\ngno = \"0.9\"\n")
	write("zz.gno", "package zz\n\nfunc F() {}\n")
	write("zz_test.gno", "package zz\n\nfunc G() {}\n")
	write("README.md", "not gno\n")
	write(".hidden.gno", "package zz\n")
	write(filepath.Join("filetests", "a_filetest.gno"), "package main\n\nfunc main() {}\n")
	// ReadMemPackage takes only _filetest.gno from this subdirectory, so a plain
	// .gno file here is never part of the package. Reporting it would fail lint
	// for a file the chain never sees; the first version of this test omitted
	// the case and let exactly that through.
	write(filepath.Join("filetests", "helper.gno"), "package helper\n")

	mpkg, err := gno.ReadMemPackage(dir, "gno.land/p/demo/zz", gno.MPAnyAll)
	require.NoError(t, err)

	var fromMemPkg []string
	for _, f := range mpkg.Files {
		if strings.HasSuffix(f.Name, ".gno") {
			fromMemPkg = append(fromMemPkg, filepath.Base(f.Name))
		}
	}
	paths, err := gno.MemPackageFilePaths(dir, "gno.land/p/demo/zz", gno.MPAnyAll)
	require.NoError(t, err)
	var fromLint []string
	for _, p := range paths {
		if strings.HasSuffix(p, ".gno") {
			fromLint = append(fromLint, filepath.Base(p))
		}
	}
	sort.Strings(fromMemPkg)
	sort.Strings(fromLint)
	require.Equal(t, fromMemPkg, fromLint,
		"MemPackageFilePaths must list exactly the .gno files ReadMemPackage puts in the package")
}
