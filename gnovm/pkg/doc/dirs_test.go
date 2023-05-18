package doc

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tNewDirs(t *testing.T) (string, *bfsDirs) {
	t.Helper()
	p, err := filepath.Abs("./testdata/dirs")
	require.NoError(t, err)
	return p, newDirs([]string{p}, nil)
}

func TestDirs_findPackage(t *testing.T) {
	abs, d := tNewDirs(t)
	tt := []struct {
		name string
		res  []bfsDir
	}{
		{"rand", []bfsDir{
			{importPath: "rand", dir: filepath.Join(abs, "rand")},
			{importPath: "crypto/rand", dir: filepath.Join(abs, "crypto/rand")},
			{importPath: "math/rand", dir: filepath.Join(abs, "math/rand")},
		}},
		{"crypto/rand", []bfsDir{
			{importPath: "crypto/rand", dir: filepath.Join(abs, "crypto/rand")},
		}},
		{"math", []bfsDir{
			{importPath: "math", dir: filepath.Join(abs, "math")},
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
	abs, d := tNewDirs(t)
	tt := []struct {
		name string
		in   string
		res  []bfsDir
	}{
		{"rand", filepath.Join(abs, "rand"), []bfsDir{
			{importPath: "rand", dir: filepath.Join(abs, "rand")},
		}},
		{"crypto/rand", filepath.Join(abs, "crypto/rand"), []bfsDir{
			{importPath: "crypto/rand", dir: filepath.Join(abs, "crypto/rand")},
		}},
		// ignored (dir name testdata), so should not return anything.
		{"crypto/testdata/rand", filepath.Join(abs, "crypto/testdata/rand"), nil},
		{"xx", filepath.Join(abs, "xx"), nil},
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
