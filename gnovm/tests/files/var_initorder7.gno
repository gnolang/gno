package main

// Tests that a blank-identifier var spec (_) runs for its side effects in
// source order before a named var that shares the same dependencies.
// GetDeclNames() returns [] for _, so it is never registered in fdeclared,
// but it still occupies its source position in the pending list.
// Init order: counter, _ (side effect only), A.

var counter int

func incr() int { counter++; return counter }

var (
	_ = incr() // runs first; counter becomes 1; contributes nothing to fdeclared
	A = incr() // runs second; counter becomes 2; A = 2
)

func main() { println(A) }

// Output:
// 2
