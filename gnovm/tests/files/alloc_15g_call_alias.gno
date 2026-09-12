// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

func f(a0, a1, a2, a3, a4, a5, a6, a7, a8, a9 Big) int { return len(a0) }

// Regression: call args were copied into a Go local before being assigned
// to the block; variant of alloc_15_slicelit_alias.
func main() {
	println(f(a, a, a, a, a, a, a, a, a, a))
}

// Error:
// allocation limit exceeded
