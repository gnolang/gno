// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

// Regression: doOpSliceLit2 variant of alloc_15_slicelit_alias.
func main() {
	s := []Big{0: a, 1: a, 2: a, 3: a, 4: a, 5: a, 6: a, 7: a, 8: a, 9: a}
	println(len(s))
}

// Error:
// allocation limit exceeded
