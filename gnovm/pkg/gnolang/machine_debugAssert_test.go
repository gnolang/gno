//go:build debugAssert

package gnolang

import (
	"strings"
	"testing"
)

// resolvePointer's StarExpr branch is the only one whose Base can name a block
// that has already left the machine's block stack: every other branch resolves
// from the live block chain (NameExpr) or bases the pointer on a non-Block
// value. Assign2 and Deref carry the same poison check, but the compound
// assignments and inc/dec write through pv.TV directly, so `*p op= v` and
// `(*p)++` would otherwise reach a recycled block unchecked.
func TestResolvePointerStarRejectsPoisonedBase(t *testing.T) {
	// Build a block, then retire it exactly as Machine.releaseBlock does.
	b := &Block{Values: make([]TypedValue, 1, blockPoolValueCap)}
	vals := b.Values[:blockPoolValueCap:blockPoolValueCap]
	clear(vals)
	*b = Block{Values: vals[:0]}
	b.poisoned = true

	b.Values = b.Values[:1]
	b.Values[0] = TypedValue{T: IntType}
	b.Values[0].SetInt(7)

	// A pointer into the retired block, as it would sit on the operand stack
	// for a `*p += 1` LHS.
	ptr := PointerValue{TV: &b.Values[0], Base: b, Index: 0}
	operand := TypedValue{T: &PointerType{Elt: IntType}, V: ptr}

	m := NewMachineWithOptions(MachineOptions{})
	defer m.Release()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("resolvePointer(*StarExpr) accepted a pointer into a recycled block")
		}
		if msg, _ := r.(string); !strings.Contains(msg, "recycled block") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	m.resolvePointer(&StarExpr{}, []TypedValue{operand})
}
