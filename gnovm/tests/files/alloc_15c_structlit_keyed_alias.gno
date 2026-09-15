// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

type S struct{ A, B, C, D, E, F, G, H, I, J Big }

// Regression: doOpStructLit (fields by name) variant of alloc_15_slicelit_alias.
func main() {
	s := S{A: a, B: a, C: a, D: a, E: a, F: a, G: a, H: a, I: a, J: a}
	println(len(s.A))
}

// Error:
// allocation limit exceeded
