// Regression for the goto interaction: bs.NumValues, captured at range init and
// used by the goto handler to restore the value stack, must EXCLUDE the
// ASSIGN-form LHS pointer operands. Otherwise a goto out of the loop body
// restores the stack one (or more) entries too tall, and the next subject read
// (which runs with the operands already consumed) reads a stale operand slot
// instead of X. With only the read-offset fix, the goto cases below silently
// persisted a wrong value (worse than the pre-fix panic).
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

	// control: name-only ASSIGN (depth 0) with goto must be unaffected.
	var a, b int
	for a, b = range []int{5, 6, 7} {
		goto N
	N:
	}
	println(a, b)
}

// Output:
// 3 40
// 3 40
// 2 7
