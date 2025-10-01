package doc

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return wd
}

func wdJoin(t *testing.T, arg string) string {
	t.Helper()
	return filepath.Join(getwd(t), arg)
}

func TestNewDirs_nonExisting(t *testing.T) {
	old := log.Default().Writer()
	var buf bytes.Buffer
	log.Default().SetOutput(&buf)
	defer func() { log.Default().SetOutput(old) }() // in case of panic

	// git doesn't track empty directories; so need to create this one on our own.
	de := wdJoin(t, "testdata/dirsempty")
	require.NoError(t, os.MkdirAll(de, 0o755))

	d := newDirs([]string{wdJoin(t, "non/existing/dir"), de}, []string{wdJoin(t, "and/this/one/neither")})
	for _, ok := d.Next(); ok; _, ok = d.Next() { //nolint:revive
	}
	log.Default().SetOutput(old)
	assert.Empty(t, d.hist, "hist should be empty")
	assert.Equal(t, strings.Count(buf.String(), "\n"), 2, "output should contain 2 lines")
	assert.Contains(t, buf.String(), "non/existing/dir: no such file or directory")
	assert.Contains(t, buf.String(), "this/one/neither/gnomod.toml: no such file or directory")
	assert.NotContains(t, buf.String(), "dirsempty: no such file or directory")
}

func TestNewDirs_invalidModDir(t *testing.T) {
	old := log.Default().Writer()
	var buf bytes.Buffer
	log.Default().SetOutput(&buf)
	defer func() { log.Default().SetOutput(old) }() // in case of panic

	d := newDirs(nil, []string{wdJoin(t, "testdata/dirs")})
	for _, ok := d.Next(); ok; _, ok = d.Next() { //nolint:revive
	}
	log.Default().SetOutput(old)
	assert.Empty(t, d.hist, "hist should be len 0 (testdata/dirs is not a valid mod dir)")
	assert.Equal(t, strings.Count(buf.String(), "\n"), 1, "output should contain 1 line")
	assert.Contains(t, buf.String(), "gnomod.toml: no such file or directory")
}

func tNewDirs(t *testing.T) (string, *bfsDirs) {
	t.Helper()

	// modify GNOHOME to testdata/dirsdep -- this allows us to test
	// dependency lookup by dirs.
	t.Setenv("GNOHOME", wdJoin(t, "testdata/dirsdep"))

	return wdJoin(t, "testdata"),
		newDirs([]string{wdJoin(t, "testdata/dirs")}, []string{wdJoin(t, "testdata/dirsmod")})
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
			{importPath: "dirs.mod/r/prefix/math/rand", dir: filepath.Join(td, "dirsmod/math/rand")},
		}},
		{"crypto/rand", []bfsDir{
			{importPath: "crypto/rand", dir: filepath.Join(td, "dirs/crypto/rand")},
		}},
		{"dep", []bfsDir{}},
		{"alpha", []bfsDir{
			{importPath: "module/alpha", dir: filepath.Join(td, "dirs/module/alpha")},
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
		t.Run("name_"+strings.ReplaceAll(tc.name, "/", "_"), func(t *testing.T) {
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
		{"2xx", "/2xx", nil},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(strings.ReplaceAll(tc.name, "/", "_"), func(t *testing.T) {
			res := d.findDir(tc.in)
			assert.Equal(t, tc.res, res, "dirs returned should be the equal")
		})
	}
}
