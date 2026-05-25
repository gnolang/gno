package gnolang

import "testing"

// TestDoOpAssign_DuplicateLHS_LeftToRight calls doOpAssign directly with
// duplicate LHS targets (all pointing to the same block slot) and verifies
// the assignments run in left-to-right order: rvs[n-1] wins. This is a
// unit-level guardrail for the bug fixed in
// "fix(gnovm): assign LHS pointers in left-to-right order for AssignStmt".
func TestDoOpAssign_DuplicateLHS_LeftToRight(t *testing.T) {
	for _, n := range []int{2, 3, 5, 17} { // 17 forces the heap-fallback branch
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
