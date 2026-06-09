package gnolang

import "testing"

// benchAssignNameN benchmarks the NameExpr LHS path (`a, a, …, a = 1, 2, …, n`).
// Operand frames are empty here (NameExpr pushes nothing for PushForPointer), so
// this isolates the multi-LHS bookkeeping overhead — the numStackValuesForPointer
// passes, the PopValues window, and the resolve+assign loop — with no per-LHS
// operand work.
func benchAssignNameN(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()
	blk, nxs := benchBlockVars(m, 1)
	lhs := make([]Expr, n)
	for i := range n {
		lhs[i] = nxs[0] // all alias the same slot; harmless for timing
	}
	stmt := &AssignStmt{Lhs: lhs, Op: ASSIGN}
	blk.Values[0] = TypedValue{T: IntType, N: i2n(0)}

	rvs := make([]TypedValue, n)
	for i := range n {
		rvs[i] = TypedValue{T: IntType, N: i2n(int64(i + 1))}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, rv := range rvs {
			m.PushValue(rv)
		}
		m.PushStmt(stmt)
		m.doOpAssign()
	}
}

func BenchmarkDoOpAssign_Name_N1(b *testing.B) { benchAssignNameN(b, 1) }
func BenchmarkDoOpAssign_Name_N2(b *testing.B) { benchAssignNameN(b, 2) }
func BenchmarkDoOpAssign_Name_N3(b *testing.B) { benchAssignNameN(b, 3) }
func BenchmarkDoOpAssign_Name_N5(b *testing.B) { benchAssignNameN(b, 5) }

// benchAssignIndexN benchmarks the IndexExpr LHS path (`s[0], s[1], … = 10, 11, …`)
// — the heaviest multi-LHS case, since each LHS contributes 2 operand-frame
// values (X and Index) that resolvePointer reads in place from the PopValues
// window.
func benchAssignIndexN(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()

	st := m.Alloc.NewType(&SliceType{Elt: IntType})
	baseArray := m.Alloc.NewListArray(nil, n)
	for i := range n {
		baseArray.List[i] = TypedValue{T: IntType, N: i2n(0)}
	}
	sv := m.Alloc.NewSlice(baseArray, 0, n, n)

	ix := &IndexExpr{}
	lhs := make([]Expr, n)
	for i := range n {
		lhs[i] = ix // shape-only; PopAsPointer reads from the value stack
	}
	stmt := &AssignStmt{Lhs: lhs, Op: ASSIGN}

	// Pre-build the push sequence once: per LHS, push X (slice) and Index;
	// then RHS values.
	pushes := make([]TypedValue, 0, 3*n)
	for i := range n {
		pushes = append(pushes,
			TypedValue{T: st, V: sv},
			TypedValue{T: IntType, N: i2n(int64(i))}, // distinct indices
		)
	}
	for i := range n {
		pushes = append(pushes, TypedValue{T: IntType, N: i2n(int64(10 + i))})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tv := range pushes {
			m.PushValue(tv)
		}
		m.PushStmt(stmt)
		m.doOpAssign()
	}
}

func BenchmarkDoOpAssign_Index_N1(b *testing.B) { benchAssignIndexN(b, 1) }
func BenchmarkDoOpAssign_Index_N2(b *testing.B) { benchAssignIndexN(b, 2) }
func BenchmarkDoOpAssign_Index_N3(b *testing.B) { benchAssignIndexN(b, 3) }
func BenchmarkDoOpAssign_Index_N5(b *testing.B) { benchAssignIndexN(b, 5) }
