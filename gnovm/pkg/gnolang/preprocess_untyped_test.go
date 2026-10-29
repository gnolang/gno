package gnolang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// preprocessVarValue preprocesses a package body and returns the preprocessed
// initializer of the package-level variable name.
func preprocessVarValue(t *testing.T, body, name string) Expr {
	t.Helper()
	m := NewMachine("", nil)
	defer m.Release()
	fn := m.MustParseFile("a.gno", "package p\n"+body)
	m.PreprocessFiles("p", "test/p", &FileSet{Files: []*FileNode{fn}}, false, false)
	for _, d := range fn.Decls {
		vd, ok := d.(*ValueDecl)
		if !ok {
			continue
		}
		for i, nx := range vd.NameExprs {
			if nx.Name == Name(name) {
				return vd.Values[i]
			}
		}
	}
	t.Fatalf("no var %q in %q", name, body)
	return nil
}

// preprocessPanic preprocesses a package body and returns the panic message.
func preprocessPanic(t *testing.T, body string) (msg string) {
	t.Helper()
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected a preprocess error for %q", body)
		msg = fmt.Sprint(r)
	}()
	preprocessVarValue(t, body, "")
	return ""
}

// requireTreeTyped asserts that every node of an untyped-shift tree carries
// the static type want and every constant leaf was converted to it.
func requireTreeTyped(t *testing.T, x Expr, want Type) {
	t.Helper()
	switch x := x.(type) {
	case *ConstExpr:
		require.Equal(t, want, x.T, "leaf %v", x)
	case *BinaryExpr:
		require.Equal(t, want, x.GetAttribute(ATTR_TYPEOF_VALUE), "node %v", x)
		requireTreeTyped(t, x.Left, want)
		if x.Op != SHL && x.Op != SHR {
			requireTreeTyped(t, x.Right, want)
		}
	case *UnaryExpr:
		require.Equal(t, want, x.GetAttribute(ATTR_TYPEOF_VALUE), "node %v", x)
		requireTreeTyped(t, x.X, want)
	default:
		t.Fatalf("unexpected node %T %v", x, x)
	}
}

func TestUntypedShift_TypedContext(t *testing.T) {
	x := preprocessVarValue(t, "var s uint\nvar x int64 = ('a' << s) + (1 << s)\n", "x")
	requireTreeTyped(t, x, Int64Type)
	y := preprocessVarValue(t, "var s uint\nvar y uint8 = ('a' << s) * (1 << s) - 1\n", "y")
	requireTreeTyped(t, y, Uint8Type)
	// The constant leaf may exceed the default type (rune) but not int64.
	z := preprocessVarValue(t, "var s uint\nvar z int64 = ('a' << s) + (1 << 40)\n", "z")
	requireTreeTyped(t, z, Int64Type)
}

func TestUntypedShift_NoContext(t *testing.T) {
	// int < rune < float: the later kind's default type.
	x := preprocessVarValue(t, "var s uint\nvar x = 1 + ('a' << s)\n", "x")
	requireTreeTyped(t, x, Int32Type)
	y := preprocessVarValue(t, "var s uint\nvar y = (1 << s) + 'a'\n", "y")
	requireTreeTyped(t, y, Int32Type)
	z := preprocessVarValue(t, "var s uint\nvar z = (1 << s) + (1 << 40)\n", "z")
	requireTreeTyped(t, z, IntType)
}

func TestUntypedShift_InterfaceContext(t *testing.T) {
	x := preprocessVarValue(t, "var s uint\nvar x interface{} = 1 + ('a' << s)\n", "x")
	requireTreeTyped(t, x, Int32Type)
	y := preprocessVarValue(t, "var s uint\nvar y any = interface{}(('a' << s) * 2)\n", "y")
	cx, ok := y.(*CallExpr)
	require.True(t, ok, "%v", y)
	requireTreeTyped(t, cx.Args[0], Int32Type)
}

func TestUntypedShift_Comparison(t *testing.T) {
	// Both operands take the common kind's default type.
	b := preprocessVarValue(t, "var s uint\nvar b = ('a' << s) < (1 << s)\n", "b")
	bx := b.(*BinaryExpr)
	requireTreeTyped(t, bx.Left, Int32Type)
	requireTreeTyped(t, bx.Right, Int32Type)
	c := preprocessVarValue(t, "var s uint\nvar c = ('a' << s) < 100\n", "c")
	cx := c.(*BinaryExpr)
	requireTreeTyped(t, cx.Left, Int32Type)
	requireTreeTyped(t, cx.Right, Int32Type)
	// A typed operand supplies the type.
	d := preprocessVarValue(t, "var s uint\nvar i int64\nvar d = ('a' << s) > i\n", "d")
	requireTreeTyped(t, d.(*BinaryExpr).Left, Int64Type)
}

func TestUntypedShift_Nested(t *testing.T) {
	x := preprocessVarValue(t, "var s uint\nvar x int64 = ((1 << s) << s) + -(1 << s)\n", "x")
	requireTreeTyped(t, x, Int64Type)
	y := preprocessVarValue(t, "var s uint\nvar y = ^(('a' << s) | 1) << s\n", "y")
	requireTreeTyped(t, y, Int32Type)
}

func TestUntypedShift_RangeCheck(t *testing.T) {
	for _, body := range []string{
		"var s uint\nvar r rune = (1 << s) + (1 << 40)\n",
		"var s uint\nvar x interface{} = ('a' << s) + (1 << 40)\n",
		"var s uint\nvar x = ('a' << s) + (1 << 40)\n",
	} {
		msg := preprocessPanic(t, body)
		require.Contains(t, msg, "bigint overflows target kind", body)
	}
	msg := preprocessPanic(t, "var s uint\nvar x = ('a' << s) + 1.5\n")
	require.Contains(t, msg, "operator << not defined on: Float64Kind")
}

func TestUntypedRuneConst_Kind(t *testing.T) {
	for _, tc := range []struct {
		expr string
		want Type
	}{
		{"'a' + 1", Int32Type},
		{"1 + 'a'", Int32Type},
		{"'a' * 2", Int32Type},
		{"'a' << 1", Int32Type},
		{"1.5 + 'a'", Float64Type},
		{"'a' + 1.5", Float64Type},
		{"1 + 1", IntType},
	} {
		x := preprocessVarValue(t, "var x = "+tc.expr+"\n", "x")
		cx, ok := x.(*ConstExpr)
		require.True(t, ok, "%s: %v", tc.expr, x)
		require.Equal(t, tc.want, cx.T, tc.expr)
	}
	c := preprocessVarValue(t, "const c = 'a' + 1\n", "c")
	require.Equal(t, UntypedRuneType, c.(*ConstExpr).T)
}

// TestPushUntypedType_GasCharged verifies the push-down walk is billed per
// node: 40 more `(1 << s) +` terms are 120 more nodes.
func TestPushUntypedType_GasCharged(t *testing.T) {
	src := func(n int) string {
		return "package pushgas\nfunc f(s uint) int64 { return " +
			strings.Repeat("(1 << s) + ", n) + "1 }\nfunc main() {}\n"
	}
	_, _, small := runPkgSource(t, "pushgas", 1<<40, src(2))
	_, _, large := runPkgSource(t, "pushgas", 1<<40, src(42))
	require.GreaterOrEqual(t, large-small, int64(120)*OpCPUSlopeCheckNode*GasFactorCPU)
}
