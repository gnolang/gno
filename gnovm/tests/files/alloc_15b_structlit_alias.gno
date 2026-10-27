// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

type S struct{ A, B, C, D, E, F, G, H, I, J Big }

// Regression: doOpStructLit (fields in order) variant of alloc_15_slicelit_alias.
func main() {
	s := S{a, a, a, a, a, a, a, a, a, a}
	println(len(s.A))
}

// Error:
// allocation limit exceeded
