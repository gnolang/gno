// Multi-value var decl where one name depends on the other via a single-value decl.
// A and B are declared together; C depends on B, and A gets C.
// This triggered a false cycle detection before the fix.
package main

var (
	A, B = C, 2
	C    = B + 1
)

func main() {
	println(A, B, C) // 3 2 3
}

// Output:
// 3 2 3
