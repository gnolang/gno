package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Equal(t, exp, v.GnoFunc)
		assert.Equal(t, exp, v.GoFunc)
		assert.Equal(t, "std", v.GnoImportPath)
		assert.Equal(t, testdataDir+"linkFunctions/std", v.GoImportPath)

		assert.Equal(t, i&machine != 0, v.MachineParam, "MachineParam should match expected value")
		if i&param != 0 {
			// require, otherwise the following would panic
			require.Len(t, v.Params, 1)
			p := v.Params[0]
			assert.Equal(t, "int", p.GnoType())
			assert.Equal(t, "int", p.GoQualifiedName())
			assert.False(t, p.IsTypedValue)
		} else {
			assert.Len(t, v.Params, 0)
		}
		if i&ret != 0 {
			// require, otherwise the following would panic
			require.Len(t, v.Results, 1)
			p := v.Results[0]
			assert.Equal(t, "int", p.GnoType())
			assert.Equal(t, "int", p.GoQualifiedName())
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

	assert.Equal(t, false, mappings[0].MachineParam)
	assert.Equal(t, "t1", mappings[0].GnoFunc)
	assert.Equal(t, "X_t1", mappings[0].GoFunc)

	assert.Equal(t, true, mappings[1].MachineParam)
	assert.Equal(t, "t2", mappings[1].GnoFunc)
	assert.Equal(t, "X_t2", mappings[1].GoFunc)
}

func Test_linkFunctions_TypedValue(t *testing.T) {
	chdir(t, "testdata/linkFunctions_TypedValue")

	pkgs, err := walkStdlibs(".")
	require.NoError(t, err)

	mappings := linkFunctions(pkgs)
	require.Len(t, mappings, 3)

	assert.Equal(t, false, mappings[0].MachineParam)
	assert.Equal(t, "TVParam", mappings[0].GnoFunc)
	assert.Equal(t, "TVParam", mappings[0].GoFunc)
	assert.Len(t, mappings[0].Results, 0)
	_ = assert.Len(t, mappings[0].Params, 1) &&
		assert.Equal(t, true, mappings[0].Params[0].IsTypedValue) &&
		assert.Equal(t, "struct{m1 map[string]interface{}}", mappings[0].Params[0].GnoType())

	assert.Equal(t, false, mappings[1].MachineParam)
	assert.Equal(t, "TVResult", mappings[1].GnoFunc)
	assert.Equal(t, "TVResult", mappings[1].GoFunc)
	assert.Len(t, mappings[1].Params, 0)
	_ = assert.Len(t, mappings[1].Results, 1) &&
		assert.Equal(t, true, mappings[1].Results[0].IsTypedValue) &&
		assert.Equal(t, "interface{S() map[int]Banker}", mappings[1].Results[0].GnoType())

	assert.Equal(t, true, mappings[2].MachineParam)
	assert.Equal(t, "TVFull", mappings[2].GnoFunc)
	assert.Equal(t, "TVFull", mappings[2].GoFunc)
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

		{"map[string]int", "map[string]int", ""},
		{"map[string][]byte", "map[string][]byte", ""},
		{"map[string]map[string]int", "map[string]map[string]int", ""},
		{"map[int]*struct{}", "map[int]*struct{}", ""},
		{"map[string]int", "map[string]string", "does not match"},
		{"map[int]string", "map[string]string", "does not match"},

		{"struct{ M int }", "struct{ M int }", ""},
		{"struct{ M int }", "struct{ M string }", "does not match"},
		{"struct{ M int }", "struct{ N int }", "does not match"},
		{"struct{ M int }", "struct{ M int; N int }", "does not match"},
		{"struct{ M int }", "struct{}", "does not match"},

		{"interface{}", "interface{}", ""},
		{"interface{ M() }", "interface{ M() }", ""},
		{"interface{ M() int }", "interface{ M() int }", ""},
		{"interface{ M() int; N() string }", "interface{ M() int; N() string }", ""},
		{"interface{ M() int }", "interface{ M() string }", "does not match"},
		{"interface{ M() int }", "interface{ M() }", "does not match"},
		{"interface{ M() int }", "interface{ N() int }", "does not match"},
		{"interface{ M() int; N() string }", "interface{ M() int; N() int }", "does not match"},

		{"func()", "func()", ""},
		{"func() int", "func() int", ""},
		{"func(s string)", "func(s string)", ""},
		// parameter name doesn't matter. only type
		{"func(s string)", "func(a string)", ""},
		{"func(s string) int", "func(s string) int", ""},
		{"func(s string) (int, int)", "func(s string) (int, int)", ""},
		{"func(s string) (int, int)", "func(s string) (p1 int, p2 int)", ""},
		{"func(s string) (int, int)", "func(s string) (p1, p2 int)", ""},
		{"func(s, v string)", "func(s string, v string)", ""},
		{"func() int", "func() string", "does not match"},
		{"func(s string) int", "func(s string) string", "does not match"},
		{"func(s string) int", "func(s string, i int) int", "does not match"},
		{"func(s string) int", "func(s string) (int, error)", "does not match"},
		{"func(s string)", "func(s string) (int, error)", "does not match"},

		{"func(s ...int)", "func(s ...int)", ""},
		{"func(s ...int)", "func(s ...float64)", "does not match"},
		{"func(s ...int)", "func(s ...int, i int)", "does not match"},

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

func Test_typesEqual_SelectorExpr(t *testing.T) {
	tests := []struct {
		name       string
		gnoExpr    *ast.SelectorExpr
		goExpr     *ast.SelectorExpr
		gnoImports []*ast.ImportSpec
		goImports  []*ast.ImportSpec
		wantErr    bool
	}{
		{
			name: "matching selector expressions with same import",
			gnoExpr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Time"},
			},
			goExpr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Time"},
			},
			gnoImports: []*ast.ImportSpec{
				{Path: &ast.BasicLit{Value: `"time"`}},
			},
			goImports: []*ast.ImportSpec{
				{Path: &ast.BasicLit{Value: `"time"`}},
			},
		},
		{
			name: "non-matching selector expressions due to different selectors",
			gnoExpr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Time"},
			},
			goExpr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Duration"},
			},
			gnoImports: []*ast.ImportSpec{
				{Path: &ast.BasicLit{Value: `"time"`}},
			},
			goImports: []*ast.ImportSpec{
				{Path: &ast.BasicLit{Value: `"time"`}},
			},
			wantErr: true,
		},
		{
			name: "matching selector expressions with different package aliases",
			gnoExpr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "t"}, // import (t "time")
				Sel: &ast.Ident{Name: "Time"},
			},
			goExpr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"}, // import "time"
				Sel: &ast.Ident{Name: "Time"},
			},
			gnoImports: []*ast.ImportSpec{
				{Name: &ast.Ident{Name: "t"}, Path: &ast.BasicLit{Value: `"time"`}},
			},
			goImports: []*ast.ImportSpec{
				{Path: &ast.BasicLit{Value: `"time"`}},
			},
		},
		{
			name: "non-matching selector expressions due to different packages",
			gnoExpr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Time"},
			},
			goExpr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "fmt"},
				Sel: &ast.Ident{Name: "Stringer"},
			},
			gnoImports: []*ast.ImportSpec{
				{Path: &ast.BasicLit{Value: `"time"`}},
			},
			goImports: []*ast.ImportSpec{
				{Path: &ast.BasicLit{Value: `"fmt"`}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mapping{
				gnoImports: tt.gnoImports,
				goImports:  tt.goImports,
			}
			err := m.typesEqual(tt.gnoExpr, tt.goExpr)
			if (err != nil) != tt.wantErr {
				t.Errorf("typesEqual() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_typesEqual_panic(t *testing.T) {
	tt := []struct {
		gnoe, goe string
		panic     string
	}{
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
