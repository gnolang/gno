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

// The directive check runs before anything parses the package, so it cannot
// take its file list from the mempackage and enumerates the directory itself.
// That duplication is only safe while the two agree: when they drifted, a
// directive in filetests/ passed lint and was then rejected at AddPackage.
func TestLintGnoFilesMatchesReadMemPackage(t *testing.T) {
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

	mpkg, err := gno.ReadMemPackage(dir, "gno.land/p/demo/zz", gno.MPAnyAll)
	require.NoError(t, err)

	var fromMemPkg []string
	for _, f := range mpkg.Files {
		if strings.HasSuffix(f.Name, ".gno") {
			fromMemPkg = append(fromMemPkg, filepath.Base(f.Name))
		}
	}
	var fromLint []string
	for _, p := range lintGnoFiles(dir) {
		fromLint = append(fromLint, filepath.Base(p))
	}
	sort.Strings(fromMemPkg)
	sort.Strings(fromLint)
	require.Equal(t, fromMemPkg, fromLint,
		"lintGnoFiles must cover exactly the .gno files ReadMemPackage puts in the package")
}
