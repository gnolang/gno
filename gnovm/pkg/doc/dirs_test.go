package doc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tNewDirs(t *testing.T) (string, *bfsDirs) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	// modify GNO_HOME to testdata/dirsdep -- this allows us to test
	// dependency lookup by dirs.
	old, ex := os.LookupEnv("GNO_HOME")
	os.Setenv("GNO_HOME", filepath.Join(wd, "testdata/dirsdep"))
	t.Cleanup(func() {
		if ex {
			os.Setenv("GNO_HOME", old)
		} else {
			os.Unsetenv("GNO_HOME")
		}
	})

	return filepath.Join(wd, "testdata"),
		newDirs([]string{filepath.Join(wd, "testdata/dirs")}, []string{filepath.Join(wd, "testdata/dirsmod")})
}

func TestDirs_findPackage(t *testing.T) {
	td, d := tNewDirs(t)
	tt := []struct {
		name string
		res  []bfsDir
	}{
		{"rand", []bfsDir{
			{importPath: "rand", dir: filepath.Join(td, "dirs/rand")},
			{importPath: "crypto/rand", dir: filepath.Join(td, "dirs/crypto/rand")},
			{importPath: "math/rand", dir: filepath.Join(td, "dirs/math/rand")},
			{importPath: "dirs.mod/prefix/math/rand", dir: filepath.Join(td, "dirsmod/math/rand")},
		}},
		{"crypto/rand", []bfsDir{
			{importPath: "crypto/rand", dir: filepath.Join(td, "dirs/crypto/rand")},
		}},
		{"dep", []bfsDir{
			{importPath: "dirs.mod/dep", dir: filepath.Join(td, "dirsdep/pkg/mod/dirs.mod/dep")},
		}},
		{"alpha", []bfsDir{
			{importPath: "dirs.mod/dep/alpha", dir: filepath.Join(td, "dirsdep/pkg/mod/dirs.mod/dep/alpha")},
			// no testdir/module/alpha as it is inside a module
		}},
		{"math", []bfsDir{
			{importPath: "math", dir: filepath.Join(td, "dirs/math")},
		}},
		{"ath", []bfsDir{}},
		{"/math", []bfsDir{}},
		{"", []bfsDir{}},
	}
	for _, tc := range tt {
		tc := tc
		t.Run("name_"+strings.Replace(tc.name, "/", "_", -1), func(t *testing.T) {
			res := d.findPackage(tc.name)
			assert.Equal(t, tc.res, res, "dirs returned should be the equal")
		})
	}
}

func TestDirs_findDir(t *testing.T) {
	td, d := tNewDirs(t)
	tt := []struct {
		name string
		in   string
		res  []bfsDir
	}{
		{"rand", filepath.Join(td, "dirs/rand"), []bfsDir{
			{importPath: "rand", dir: filepath.Join(td, "dirs/rand")},
		}},
		{"crypto/rand", filepath.Join(td, "dirs/crypto/rand"), []bfsDir{
			{importPath: "crypto/rand", dir: filepath.Join(td, "dirs/crypto/rand")},
		}},
		// ignored (dir name testdata), so should not return anything.
		{"crypto/testdata/rand", filepath.Join(td, "dirs/crypto/testdata/rand"), nil},
		{"xx", filepath.Join(td, "dirs/xx"), nil},
		{"xx2", "/xx2", nil},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(strings.Replace(tc.name, "/", "_", -1), func(t *testing.T) {
			res := d.findDir(tc.in)
			assert.Equal(t, tc.res, res, "dirs returned should be the equal")
		})
	}
}
