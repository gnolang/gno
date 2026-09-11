// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

func f(xs ...Big) int { return len(xs) }

// Regression: popCopyArgs built the variadic slice in a Go local after
// popping the args; variant of alloc_15_slicelit_alias.
func main() {
	println(f(a, a, a, a, a, a, a, a, a, a))
}

// Error:
// allocation limit exceeded
