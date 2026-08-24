package main

// Tests chained dependencies where the source order is the reverse of the
// initialization order.
// C depends on B, B depends on A; source order is C, B, A.
// Init order must be: A (no deps), B (deps A), C (deps B).

var (
	C = B + 2
	B = A + 1
	A = 1
)

func main() { println(A, B, C) }

// Output:
// 1 2 4
