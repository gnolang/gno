// Regression: ASSIGN-form range with SelectorExpr / StarExpr LHS. Pre-fix these
// ranged over the LHS pointer operand (the struct value, resp. the pointer) and
// threw an uncatchable host panic "unexpected type for len(): main.S" / "*string".
package main

type S struct{ F string }

func main() {
	var s S
	var i int
	for i, s.F = range []string{"A", "B", "C"} {
	}
	println(i, s.F)

	x := "orig"
	p := &x
	var j int
	for j, *p = range []string{"A", "B", "C"} {
	}
	println(j, *p)
}

// Output:
// 2 C
// 2 C
