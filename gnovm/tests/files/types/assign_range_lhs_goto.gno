// Regression for the goto interaction: bs.NumValues, captured at range init and
// used by the goto handler to restore the value stack, must EXCLUDE the
// ASSIGN-form LHS pointer operands. Otherwise a goto out of the loop body
// restores the stack one (or more) entries too tall, and the next subject read
// (which runs with the operands already consumed) reads a stale operand slot
// instead of X. With only the read-offset fix, the goto cases below silently
// persisted a wrong value (worse than the pre-fix panic).
//
// All three range handlers capture bs.NumValues, so each is covered here:
// OpRangeIter (slice/array), OpRangeIterArrayPtr, OpRangeIterString and
// OpRangeIterMap. IndexExpr targets are included so the capture is exercised
// with more than one stack entry per target (NameExpr 0, Selector/Star 1,
// IndexExpr 2).
package main

type S struct{ F int }

func main() {
	var st S
	var k int
	for k, st.F = range []int{10, 20, 30, 40} {
		_ = []int{77, 88, 99, 111}
		goto L
	L:
	}
	println(k, st.F)

	x := -1
	p := &x
	var k2 int
	for k2, *p = range []int{10, 20, 30, 40} {
		_ = []int{77, 88, 99, 111}
		goto M
	M:
	}
	println(k2, x)

	// OpRangeIterString, IndexExpr value target (2 entries above X).
	sm := map[string]rune{"k": '?'}
	var k3 int
	for k3, sm["k"] = range "XYZ" {
		_ = []int{77, 88, 99, 111}
		goto N
	N:
	}
	println(k3, int(sm["k"]))

	// OpRangeIterMap, IndexExpr key AND value targets (4 entries above X).
	src := map[int]string{7: "hello"}
	km := map[string]int{"k": -1}
	vm := map[string]string{"k": "?"}
	for km["k"], vm["k"] = range src {
		_ = []int{77, 88, 99, 111}
		goto O
	O:
	}
	println(km["k"], vm["k"], len(vm["k"]))

	// OpRangeIterArrayPtr: the one handler that WRITES through the subject
	// slot (*xv = *dv), so a wrong index corrupts rather than only misreads.
	arr := [4]int{10, 20, 30, 40}
	pa := &arr
	am := map[string]int{"k": -1}
	var st2 S
	for am["k"], st2.F = range pa {
		_ = []int{77, 88, 99, 111}
		goto P
	P:
	}
	println(am["k"], st2.F)

	// control: name-only ASSIGN (no operands above X) with goto must be
	// unaffected.
	var a, b int
	for a, b = range []int{5, 6, 7} {
		goto Q
	Q:
	}
	println(a, b)
}

// Output:
// 3 40
// 3 40
// 2 90
// 7 hello 5
// 3 40
// 2 7
