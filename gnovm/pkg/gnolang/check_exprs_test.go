package gnolang

import (
	"strings"
	"testing"

	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// Expression rules (check_exprs.go) through accumulate mode: one program
// per rule, the expected message or none.
func TestPreprocessCollectExprRules(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		// `case nil` is how Go compares a slice, map or func tag; it must
		// survive the conversion of the case to the tag type.
		{"switchSliceNil", "var s []int; switch s { case nil: }", ""},
		{"switchMapNil", "var m map[int]int; switch m { case nil: default: }", ""},
		{"switchFuncNil", "var f func(); switch f { case nil: }", ""},
		{"switchTypedNilTag", "type S []int; switch S(nil) { case nil: }", ""},
		{"switchSliceCase", "var s []int; switch s { case s: }", "[]int can only be compared to nil"},
		{"switchTypedNilCase", "type S []int; var s S; switch s { case S(nil): }", "S can only be compared to nil"},
		{"switchIfaceSliceCase", "var s []int; var i any; switch i { case s: }", "[]int can only be compared to nil"},
		{"switchNoCases", "type Q struct{ v []int }; var q Q; switch q { }", "Q is not comparable)"},
		{"switchTagless", "x := 1; switch { case x > 0: }", ""},
		// Literal indices against the array length; [...]T is measured.
		{"arrKeyOver", "_ = [3]int{5: 1}", "invalid argument: index 5 out of bounds [0:3]"},
		{"arrSeqOver", "_ = [2]int{1, 2, 3}", "index 2 is out of bounds (>= 2)"},
		{"arrKeyThenSeqOver", "_ = [2]int{1: 1, 2}", "index 2 is out of bounds (>= 2)"},
		{"arrLast", "_ = [3]int{2: 1}", ""},
		{"arrEllipsis", "_ = [...]int{5: 1}", ""},
		{"sliceKey", "_ = []int{5: 1}", ""},
		{"arrDup", "_ = [3]int{1: 1, 1: 2}", "duplicate index 1 in array or slice literal"},
		// A string has no 3-index slice; a slice does.
		{"str3", "s := \"abc\"; _ = s[1:2:3]", "invalid operation: 3-index slice of string"},
		{"constStr3", "const s = \"abc\"; _ = s[1:2:3]", "invalid operation: 3-index slice of string"},
		{"slice3", "s := []int{1, 2, 3}; _ = s[1:2:3]", ""},
	} {
		msgs := collectErrors(t, "package main\n\nfunc main() { "+tc.body+" }\n")
		got := strings.Join(msgs, "|")
		if tc.want == "" && got != "" || tc.want != "" && !strings.Contains(got, tc.want) {
			t.Errorf("%s: want %q, got %q", tc.name, tc.want, got)
		}
	}
}

// TestPreprocessGas_LiteralIndices pins the literal index check's charge:
// OpCPUSlopeCheckNode per element, nothing else.
func TestPreprocessGas_LiteralIndices(t *testing.T) {
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", `package main

var S = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

func main() { println(S) }
`)
	require.NoError(t, err)
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	PredefineFileSet(m.Store, pn, fset)
	Preprocess(m.Store, pn, fn)
	lit := fn.Decls[0].(*ValueDecl).Values[0].(*CompositeLitExpr)
	require.Len(t, lit.Elts, 12)
	gm := stypes.NewGasMeter(1_000_000_000)
	assertLiteralIndices(gm, lit, -1)
	require.Equal(t, int64(12)*OpCPUSlopeCheckNode*GasFactorCPU, gm.GasConsumed())
}
