package main

import (
	"errors"
	"fmt"
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
		{
			// vm/qfile serves a package's files sorted, and ReadMemPackage
			// flattens filetests/ into the same list, so a filetest routinely
			// sorts ahead of the production files. Its package clause is
			// arbitrary and unrelated to the package's own.
			name: "ignores a filetest sorting ahead of the production files",
			files: []*std.MemFile{
				{Name: "caller_teller_sub_realm_filetest.gno", Body: "package grc20subrealm\n"},
				{Name: "token.gno", Body: "package grc20\n\nfunc F() {}"},
			},
			want: "grc20",
		},
		{
			name: "strips the _test suffix an external test file carries",
			files: []*std.MemFile{
				{Name: "grc20_test.gno", Body: "package grc20_test\n"},
				{Name: "token.gno", Body: "package grc20\n"},
			},
			want: "grc20",
		},
		{
			name: "filetests agreeing on a clause name the package",
			files: []*std.MemFile{
				{Name: "a_filetest.gno", Body: "package anything\n"},
				{Name: "b_filetest.gno", Body: "package anything\n"},
			},
			want: "anything",
		},
		{
			name: "filetests disagreeing fall back to the chain's default",
			files: []*std.MemFile{
				{Name: "a_filetest.gno", Body: "package anything\n"},
				{Name: "b_filetest.gno", Body: "package something\n"},
			},
			want: "filetests",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, packageName(tt.files))
		})
	}
}

// TestRPCGetterNamesReconstructedPackage pins the name a dependency rebuilt from
// vm/qfile carries. The listing is what a node serves for a package holding
// filetests: production files, stored _test.gno, and flattened _filetest.gno,
// sorted, so the filetest comes first and its package clause is not the
// package's own.
func TestRPCGetterNamesReconstructedPackage(t *testing.T) {
	const pkgPath = "gno.land/p/demo/tokens/grc20"
	files := map[string]string{
		pkgPath: "caller_teller_sub_realm_filetest.gno\ngnomod.toml\ntoken.gno\ntoken_test.gno",
		path.Join(pkgPath, "caller_teller_sub_realm_filetest.gno"): "package grc20subrealm\n",
		path.Join(pkgPath, "gnomod.toml"):                          "module = \"" + pkgPath + "\"\n",
		path.Join(pkgPath, "token.gno"):                            "package grc20\n\nfunc F() {}\n",
		path.Join(pkgPath, "token_test.gno"):                       "package grc20_test\n",
	}

	g := &rpcGetter{
		cache: make(map[string]*std.MemPackage),
		qfile: func(fp string) ([]byte, error) {
			body, ok := files[fp]
			if !ok {
				return nil, errors.New("file is not available")
			}
			return []byte(body), nil
		},
	}

	mpkg := g.GetMemPackage(pkgPath)
	require.NotNil(t, mpkg)
	assert.Equal(t, "grc20", mpkg.Name,
		"the typechecker emits .gnobuiltins.gno under this name, so a filetest's "+
			"clause here rejects the importing package for a name its author never wrote")
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

// TestRPCGetterSeparatesTransportFaultsFromAbsence pins the two kinds of
// failure fetch used to collapse. "The chain answered: nothing at this path"
// is evidence about the import; "the chain could not be asked" is evidence
// about the operator's network -- and only the first may become a verdict.
func TestRPCGetterSeparatesTransportFaultsFromAbsence(t *testing.T) {
	t.Run("a fault the closure marked is remembered as no answer", func(t *testing.T) {
		g := &rpcGetter{
			cache: map[string]*std.MemPackage{},
			qfile: func(string) ([]byte, error) {
				return nil, fmt.Errorf("%w: connection refused", errResolverUnavailable)
			},
		}
		require.Nil(t, g.GetMemPackage("gno.land/p/x"))
		require.ErrorIs(t, g.transportErr, errResolverUnavailable,
			"the getter must remember that its answer was no answer at all")
	})

	t.Run("a chain that answered 'not found' is a miss, not a fault", func(t *testing.T) {
		g := &rpcGetter{
			cache: map[string]*std.MemPackage{},
			qfile: func(string) ([]byte, error) {
				return nil, errors.New("package is not available") // the node ran the query
			},
		}
		require.Nil(t, g.GetMemPackage("gno.land/p/x"))
		require.NoError(t, g.transportErr,
			"an import the chain genuinely lacks IS evidence about the package")
	})
}
