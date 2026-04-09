package main

// Tests that a local variable shadowing a package-level name does NOT
// create a false dependency. f() declares a local B that shadows the
// package-level B; the NameExpr for the local B resolves to f's block,
// not the package block, so A should have no dependency on B.
// Init order: A and B are independent; A is declared first so it runs first.

func f() int {
	B := 99
	return B
}

var (
	A = f()
	B = 42
)

func main() { println(A, B) }

// Output:
// 99 42
