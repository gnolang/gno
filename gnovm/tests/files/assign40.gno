// Verifies left-to-right assignment for duplicate LHS in AssignStmt
// (Go spec §Assignments: "the assignments are carried out in left-to-right
// order"). Each shape exercises a different branch of PopAsPointer.
package main

type T struct{ f int }

var pkgA int

func nameExpr() {
	pkgA, pkgA, pkgA = 1, 2, 3
	if pkgA != 3 {
		panic(pkgA)
	}
}

func indexSlice() {
	s := []int{0, 0}
	s[0], s[0], s[0] = 1, 2, 3
	if s[0] != 3 {
		panic(s[0])
	}
}

func indexMap() {
	m := map[string]int{}
	m["k"], m["k"], m["k"] = 1, 2, 3
	if m["k"] != 3 {
		panic(m["k"])
	}
}

func starExpr() {
	x := 0
	p := &x
	*p, *p, *p = 1, 2, 3
	if x != 3 {
		panic(x)
	}
}

func selectorExpr() {
	s := &T{}
	s.f, s.f, s.f = 1, 2, 3
	if s.f != 3 {
		panic(s.f)
	}
}

func main() {
	nameExpr()
	indexSlice()
	indexMap()
	starExpr()
	selectorExpr()
}

// Output:
