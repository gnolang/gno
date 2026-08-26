package main

import (
	"errors"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestPackageName(t *testing.T) {
	tests := []struct {
		name  string
		files []*std.MemFile
		want  string
	}{
		{
			name: "derives from first gno file",
			files: []*std.MemFile{
				{Name: "gnomod.toml", Body: "module = \"gno.land/r/x\"\n"},
				{Name: "x.gno", Body: "package x\n\nfunc F() {}"},
			},
			want: "x",
		},
		{
			name: "skips non-gno files",
			files: []*std.MemFile{
				{Name: "README.md", Body: "# not gno"},
				{Name: "foo.gno", Body: "package foo"},
			},
			want: "foo",
		},
		{
			name:  "no gno files yields empty",
			files: []*std.MemFile{{Name: "gnomod.toml", Body: "module = \"x\""}},
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, packageName(tt.files))
		})
	}
}

// TestRPCGetterCaching verifies the getter caches successful fetches (immutable
// on-chain packages) but not misses (a package absent now may appear later).
func TestRPCGetterCaching(t *testing.T) {
	const pkgPath = "gno.land/p/x"
	files := map[string]string{
		pkgPath:                     "x.gno",       // file list
		path.Join(pkgPath, "x.gno"): "package x\n", // file body
	}

	available := false
	calls := 0
	g := &rpcGetter{
		cache: make(map[string]*std.MemPackage),
		qfile: func(fp string) ([]byte, error) {
			calls++
			if !available {
				return nil, errors.New("package is not available")
			}
			body, ok := files[fp]
			if !ok {
				return nil, errors.New("file is not available")
			}
			return []byte(body), nil
		},
	}

	// Absent now: miss, and it must NOT be cached.
	assert.Nil(t, g.GetMemPackage(pkgPath))
	assert.Nil(t, g.GetMemPackage(pkgPath))

	// The package is enabled later in the run: the miss was not pinned, so it
	// now resolves.
	available = true
	mpkg := g.GetMemPackage(pkgPath)
	require.NotNil(t, mpkg, "package must resolve once available (miss not cached)")
	assert.Equal(t, "x", mpkg.Name)

	// Subsequent lookups are served from cache — no further queries.
	callsAfterHit := calls
	g.GetMemPackage(pkgPath)
	assert.Equal(t, callsAfterHit, calls, "cached package must not be re-queried")
}
