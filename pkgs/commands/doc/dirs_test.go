package doc

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDirs(t *testing.T) (string, *Dirs) {
	t.Helper()
	p, err := filepath.Abs("./testdata/dirs")
	require.NoError(t, err)
	return p, NewDirs(p)
}

func TestDirs_findPackage(t *testing.T) {
	abs, d := newDirs(t)
	tt := []struct {
		name string
		res  []Dir
	}{
		{"rand", []Dir{
			{importPath: "rand", dir: filepath.Join(abs, "rand")},
			{importPath: "crypto/rand", dir: filepath.Join(abs, "crypto/rand")},
			{importPath: "math/rand", dir: filepath.Join(abs, "math/rand")},
		}},
		{"crypto/rand", []Dir{
			{importPath: "crypto/rand", dir: filepath.Join(abs, "crypto/rand")},
		}},
		{"math", []Dir{
			{importPath: "math", dir: filepath.Join(abs, "math")},
		}},
		{"ath", []Dir{}},
		{"/math", []Dir{}},
		{"", []Dir{}},
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
	abs, d := newDirs(t)
	tt := []struct {
		name string
		in   string
		res  []Dir
	}{
		{"rand", filepath.Join(abs, "rand"), []Dir{
			{importPath: "rand", dir: filepath.Join(abs, "rand")},
		}},
		{"crypto/rand", filepath.Join(abs, "crypto/rand"), []Dir{
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
