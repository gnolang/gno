package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/jaekwon/testify/assert"
	"github.com/jaekwon/testify/require"
)

const testdataDir = "github.com/gnolang/gno/misc/genstd/testdata/"

var initWD = func() string {
	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return d
}()

func chdir(t *testing.T, s string) {
	t.Helper()

	os.Chdir(filepath.Join(initWD, s))
	t.Cleanup(func() {
		os.Chdir(initWD)
		dirsOnce = sync.Once{}
		memoGitRoot, memoRelPath = "", ""
	})
}

func Test_linkFunctions(t *testing.T) {
	chdir(t, "testdata/linkFunctions")

	pkgs, err := walkStdlibs(".")
	require.NoError(t, err)

	mappings := linkFunctions(pkgs)
	require.Len(t, mappings, 8)

	const (
		ret = 1 << iota
		param
		machine
	)
	str := func(i int) string {
		s := "Fn"
		if i&machine != 0 {
			s += "Machine"
		}
		if i&param != 0 {
			s += "Param"
		}
		if i&ret != 0 {
			s += "Ret"
		}
		return s
	}

	for i, v := range mappings {
		exp := str(i)
		assert.Equal(t, v.GnoFunc, exp)
		assert.Equal(t, v.GoFunc, exp)
		assert.Equal(t, v.GnoImportPath, "std")
		assert.Equal(t, v.GoImportPath, testdataDir+"linkFunctions/std")

		assert.Equal(t, v.MachineParam, i&machine != 0, "MachineParam should match expected value")
		if i&param != 0 {
			// require, otherwise the following would panic
			require.Len(t, v.Params, 1)
			p := v.Params[0]
			assert.Equal(t, p.GnoType(), "int")
			assert.Equal(t, p.GoQualifiedName(), "int")
			assert.False(t, p.IsTypedValue)
		} else {
			assert.Len(t, v.Params, 0)
		}
		if i&ret != 0 {
			// require, otherwise the following would panic
			require.Len(t, v.Results, 1)
			p := v.Results[0]
			assert.Equal(t, p.GnoType(), "int")
			assert.Equal(t, p.GoQualifiedName(), "int")
			assert.False(t, p.IsTypedValue)
		} else {
			assert.Len(t, v.Results, 0)
		}
	}
}

func Test_linkFunctions_unexp(t *testing.T) {
	chdir(t, "testdata/linkFunctions_unexp")

	pkgs, err := walkStdlibs(".")
	require.NoError(t, err)

	mappings := linkFunctions(pkgs)
	require.Len(t, mappings, 2)

	assert.Equal(t, mappings[0].MachineParam, false)
	assert.Equal(t, mappings[0].GnoFunc, "t1")
	assert.Equal(t, mappings[0].GoFunc, "X_t1")

	assert.Equal(t, mappings[1].MachineParam, true)
	assert.Equal(t, mappings[1].GnoFunc, "t2")
	assert.Equal(t, mappings[1].GoFunc, "X_t2")
}

func Test_linkFunctions_TypedValue(t *testing.T) {
	chdir(t, "testdata/linkFunctions_TypedValue")

	pkgs, err := walkStdlibs(".")
	require.NoError(t, err)

	mappings := linkFunctions(pkgs)
	require.Len(t, mappings, 3)

	assert.Equal(t, mappings[0].MachineParam, false)
	assert.Equal(t, mappings[0].GnoFunc, "TVParam")
	assert.Equal(t, mappings[0].GoFunc, "TVParam")
	assert.Len(t, mappings[0].Results, 0)
	_ = assert.Len(t, mappings[0].Params, 1) &&
		assert.Equal(t, mappings[0].Params[0].IsTypedValue, true) &&
		assert.Equal(t, mappings[0].Params[0].GnoType(), "struct{m1 map[string]interface{}}")

	assert.Equal(t, mappings[1].MachineParam, false)
	assert.Equal(t, mappings[1].GnoFunc, "TVResult")
	assert.Equal(t, mappings[1].GoFunc, "TVResult")
	assert.Len(t, mappings[1].Params, 0)
	_ = assert.Len(t, mappings[1].Results, 1) &&
		assert.Equal(t, mappings[1].Results[0].IsTypedValue, true) &&
		assert.Equal(t, mappings[1].Results[0].GnoType(), "interface{S() map[int]Banker}")

	assert.Equal(t, mappings[2].MachineParam, true)
	assert.Equal(t, mappings[2].GnoFunc, "TVFull")
	assert.Equal(t, mappings[2].GoFunc, "TVFull")
	assert.Len(t, mappings[2].Params, 1)
	assert.Len(t, mappings[2].Results, 1)
}

func Test_linkFunctions_noMatch(t *testing.T) {
	chdir(t, "testdata/linkFunctions_noMatch")

	pkgs, err := walkStdlibs(".")
	require.NoError(t, err)

	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Contains(t, fmt.Sprint(r), "no matching go function declaration")
	}()

	linkFunctions(pkgs)
}

func Test_linkFunctions_noMatchSig(t *testing.T) {
	chdir(t, "testdata/linkFunctions_noMatchSig")

	pkgs, err := walkStdlibs(".")
	require.NoError(t, err)

	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Contains(t, fmt.Sprint(r), "doesn't match signature of go function")
	}()

	linkFunctions(pkgs)
}

// typesEqual - separate tests.

var typesEqualMapping = &mapping{
	GnoImportPath: "std",
	GnoFunc:       "Fn",
	GoImportPath:  "github.com/gnolang/gno/gnovm/stdlibs/std",
	GoFunc:        "Fn",
	goImports: []*ast.ImportSpec{
		{
			Name: &ast.Ident{Name: "gno"},
			Path: &ast.BasicLit{Value: `"github.com/gnolang/gno/gnovm/pkg/gnolang"`},
		},
		{
			Path: &ast.BasicLit{Value: `"github.com/gnolang/gno/tm2/pkg/crypto"`},
		},
	},
	gnoImports: []*ast.ImportSpec{
		{
			Path: &ast.BasicLit{Value: `"std"`},
		},
		{
			Path: &ast.BasicLit{Value: `"math"`},
		},
	},
}

func Test_typesEqual(t *testing.T) {
	tt := []struct {
		gnoe, goe   string
		errContains string
	}{
		{"int", "int", ""},
		{"*[11][]rune", "*[11][ ]rune", ""},

		{"madeup", "madeup", "non-builtin type"},

		{"int", "string", "does not match"},
		{"*int", "int", "does not match"},
		{"string", "*string", "does not match"},
		{"*string", "*int", "does not match"},

		{"[]int", "[1]int", "does not match"},
		{"[1]int", "[]int", "does not match"},
		{"[2]int", "[2]string", "does not match"},
		// valid, but unsupported (only BasicLits)
		{"[(11)]int", "[(11)]string", "does not match"},

		// even though mathematically equal, for simplicity we don't implement
		// "true" basic lit equivalence
		{"[8]int", "[0x8]int", "does not match"},
	}

	for idx, tv := range tt {
		t.Run(fmt.Sprintf("%02d_%s", idx, tv.gnoe), func(t *testing.T) {
			gnoe, err := parser.ParseExpr(tv.gnoe)
			require.NoError(t, err)
			goe, err := parser.ParseExpr(tv.goe)
			require.NoError(t, err)

			err = typesEqualMapping.typesEqual(gnoe, goe)
			if tv.errContains == "" {
				assert.NoError(t, err)
			} else {
				_ = assert.Error(t, err) &&
					assert.Contains(t, err.Error(), tv.errContains)
			}
		})
	}
}

func Test_typesEqual_panic(t *testing.T) {
	tt := []struct {
		gnoe, goe string
		panic     string
	}{
		{"map[string]string", "map[string]string", "not implemented"},
		{"func(s string)", "func(s string)", "not implemented"},
		{"interface{}", "interface{}", "not implemented"},
		{"struct{}", "struct{}", "not implemented"},
		{"1 + 2", "1 + 2", "invalid expression"},
	}

	for _, tv := range tt {
		t.Run(tv.gnoe, func(t *testing.T) {
			gnoe, err := parser.ParseExpr(tv.gnoe)
			require.NoError(t, err)
			goe, err := parser.ParseExpr(tv.goe)
			require.NoError(t, err)

			defer func() {
				r := recover()
				if tv.panic == "" {
					assert.Nil(t, r)
				} else {
					assert.Contains(t, fmt.Sprint(r), tv.panic)
				}
			}()

			result := typesEqualMapping.typesEqual(gnoe, goe)
			assert.Nil(t, result)
		})
	}
}
