package gnolang

import "testing"

// TestDoOpAssign_DuplicateNameLHS_LeftToRight: `a, a, …, a = 1, 2, …, n` —
// NameExpr LHS all aliasing the same block slot. Covers the forward-Assign2
// half of the fix: the rightmost RHS (rvs[n-1] == n) wins.
// (NameExpr's PopAsPointer consumes no value-stack sub-evals, so this case
// does not exercise per-LHS operand-frame addressing — that's what
// TestDoOpAssign_DistinctIndexLHS_InPlaceFrames covers.)
func TestDoOpAssign_DuplicateNameLHS_LeftToRight(t *testing.T) {
	for _, n := range []int{2, 3, 5, 17} {
		m := benchMachine()
		blk, nxs := benchBlockVars(m, 1)
		lhs := make([]Expr, n)
		for i := range n {
			lhs[i] = nxs[0] // every LHS points to the SAME slot
		}
		stmt := &AssignStmt{Lhs: lhs, Op: ASSIGN}

		blk.Values[0] = TypedValue{T: IntType, N: i2n(0)}
		for i := range n {
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i + 1))}) // rvs = 1,2,...,n
		}
		m.PushStmt(stmt)
		m.doOpAssign()

		if got, want := blk.Values[0].GetInt(), int64(n); got != want {
			t.Errorf("n=%d: got %d, want %d (rightmost RHS should win)", n, got, want)
		}
		m.Release()
	}
}

// TestDoOpAssign_DistinctIndexLHS_InPlaceFrames: `s[0], s[1] = 10, 20` —
// IndexExpr LHS with distinct indices on the same slice. doOpAssign resolves
// each LHS forward (left to right) from its own operand frame, read in place
// from the value-stack window. If the per-LHS offset arithmetic were wrong
// (e.g. frames addressed in the wrong order), lvs[0] would resolve to slice[1]
// and vice versa, swapping the final element values.
func TestDoOpAssign_DistinctIndexLHS_InPlaceFrames(t *testing.T) {
	m := benchMachine()
	defer m.Release()

	st := m.Alloc.NewType(&SliceType{Elt: IntType})
	baseArray := m.Alloc.NewListArray(nil, 2)
	baseArray.List[0] = TypedValue{T: IntType, N: i2n(0)}
	baseArray.List[1] = TypedValue{T: IntType, N: i2n(0)}
	sv := m.Alloc.NewSlice(baseArray, 0, 2, 2)

	ix := &IndexExpr{} // shape-only placeholder; PopAsPointer reads from the stack, not the AST
	stmt := &AssignStmt{Lhs: []Expr{ix, ix}, Op: ASSIGN}

	// Production push order (per op_exec.go AssignStmt case + PushForPointer):
	// Lhs[0].X, Lhs[0].Index, Lhs[1].X, Lhs[1].Index, Rhs[0], Rhs[1].
	m.PushValue(TypedValue{T: st, V: sv})           // Lhs[0].X
	m.PushValue(TypedValue{T: IntType, N: i2n(0)})  // Lhs[0].Index
	m.PushValue(TypedValue{T: st, V: sv})           // Lhs[1].X
	m.PushValue(TypedValue{T: IntType, N: i2n(1)})  // Lhs[1].Index
	m.PushValue(TypedValue{T: IntType, N: i2n(10)}) // Rhs[0]
	m.PushValue(TypedValue{T: IntType, N: i2n(20)}) // Rhs[1]
	m.PushStmt(stmt)
	m.doOpAssign()

	if got := baseArray.List[0].GetInt(); got != 10 {
		t.Errorf("slice[0]: got %d, want 10 (lvs[0] should resolve to index 0)", got)
	}
	if got := baseArray.List[1].GetInt(); got != 20 {
		t.Errorf("slice[1]: got %d, want 20 (lvs[1] should resolve to index 1)", got)
	}
}
