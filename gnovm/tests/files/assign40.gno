// Verifies that in an AssignStmt with duplicate LHS targets referring to the
// same variable, assignments proceed in left-to-right order (Go spec
// §Assignments): a, a, a = 1, 2, 3 must leave a == 3, not a == 1.
package main

var a int

func F() {
	a, a, a = 1, 2, 3
}

func main() {
	F()
	if a != 3 {
		panic(a)
	}
	println("ok")
}

// Output:
// ok
