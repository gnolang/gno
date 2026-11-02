package gnolang

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func typeCheckInfo(t *testing.T, files map[string]string) (*types.Info, *token.FileSet) {
	t.Helper()
	mpkg := &std.MemPackage{Type: MPUserProd, Name: "p", Path: "gno.land/p/test/p"}
	mpkg.SetFile("gnomod.toml", GenGnoModLatest(mpkg.Path))
	for name, body := range files {
		mpkg.SetFile(name, body)
	}
	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
		Uses:  map[*ast.Ident]types.Object{},
	}
	fset := token.NewFileSet()
	_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{Mode: TCLatestRelaxed, ProdOnly: true, Fset: fset, Info: info})
	require.NoError(t, err)
	return info, fset
}

// TestFiles256 checks that 256-bit use is found from types, not spelling.
func TestFiles256(t *testing.T) {
	info, fset := typeCheckInfo(t, map[string]string{
		"field.gno": "package p\ntype S struct{ uint256 int }\nfunc F(s S) int { return s.uint256 }\n",
		"plain.gno": "package p\nfunc G() int { return 1 }\n",
		"value.gno": "package p\nfunc H() bool { return uint256(1) == 1 }\n",
		"named.gno": "package p\ntype Amount uint256\n",
		"slice.gno": "package p\nvar xs []int256\n",
		"user.gno":  "package p\nfunc Z() Amount { return Amount(0) }\n",
	})
	// user.gno only names Amount, a named type declared elsewhere, but
	// converts the constant 0 to it: the expression's type is Amount, not a
	// 256-bit shim, so it is not flagged by itself; the declaration is.
	assert.Equal(t, []string{"named.gno", "slice.gno", "value.gno"}, Files256(info, fset))

	refs := RefsInto256Files(info, fset, map[string]bool{"named.gno": true, "slice.gno": true, "value.gno": true})
	assert.Equal(t, []string{"user.gno: Amount"}, refs)
	assert.Empty(t, RefsInto256Files(info, fset, map[string]bool{"value.gno": true}))
}

func TestInitEffects256(t *testing.T) {
	src := "package p\nvar bits uint256 = 1 << 255\nvar min = int256(bits)\nvar n = count()\nfunc count() int { return 1 }\nfunc init() { n++ }\n"
	info, fset := typeCheckInfo(t, map[string]string{"a.gno": src})
	effects, err := InitEffects256(info, fset, "a.gno", src)
	require.NoError(t, err)
	assert.Equal(t, []string{"call in a variable initializer at line 4", "func init"}, effects)

	clean := "package p\nvar bits uint256 = 1 << 255\nvar min = int256(bits)\nfunc F() int { return count() }\nfunc count() int { return 1 }\n"
	info, fset = typeCheckInfo(t, map[string]string{"a.gno": clean})
	effects, err = InitEffects256(info, fset, "a.gno", clean)
	require.NoError(t, err)
	assert.Empty(t, effects)
}
