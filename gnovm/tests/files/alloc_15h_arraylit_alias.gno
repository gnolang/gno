// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

// NOT a regression test for doOpArrayLit's anchor, despite the shape: it
// passes with that anchor removed. defaultArrayValue allocates [10]Big at
// full size (3MB) before the copy loop runs, so the 1MB cap trips at
// construction and the anchored window is never reached. Kept as a cap-
// enforcement case for the array literal; the anchor itself is defence in
// depth: for a value element type a copy is the same size as the default it
// replaces, so the exposure is ~2x rather than N x, and no discriminating
// shape is known.
func main() {
	s := [10]Big{a, a, a, a, a, a, a, a, a, a}
	println(len(s))
}

// Error:
// allocation limit exceeded
