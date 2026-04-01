package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShiftExprAttrTypeOfValue(t *testing.T) {
	t.Parallel()

	m := NewMachine("test", nil)
	c := `package test
func main() {
	var a uint = 1
	b := make([]byte, 1<<a)
	println(b)
}`
	n := m.MustParseFile("main.go", c)
	m.RunFiles(n)

	fn := n.Decls[0].(*FuncDecl)
	assignStmt := fn.Body[1].(*AssignStmt)
	callExpr := assignStmt.Rhs[0].(*CallExpr)

	// The shift expression (1<<a) is the second argument to make.
	bx, ok := callExpr.Args[1].(*BinaryExpr)
	require.True(t, ok, "expected BinaryExpr for shift")
	assert.Equal(t, SHL, bx.Op)

	// When checkOrConvertType processes the shift expression, it converts
	// bx.Left (the literal 1) from UntypedBigintType to IntType. Without
	// explicitly setting ATTR_TYPEOF_VALUE on bx itself, the shift expression
	// retains the stale UntypedBigintType from initial preprocessing.
	attr := bx.GetAttribute(ATTR_TYPEOF_VALUE)
	assert.Equal(t, IntType, attr)
}
