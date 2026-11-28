// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// doOpArrayLit copies each element given into a list that is a Go local
// until the literal is pushed, so only its anchor counts the copies already
// made when a later copy triggers a collection. Indexing the literal copies
// one element out, so no other copy of the array trips the cap first.
type Big [300000]byte

var a Big

func main() {
	x := [10]Big{a, a, a, a, a, a, a, a, a, a}[9]
	println(len(x))
}

// Error:
// allocation limit exceeded
