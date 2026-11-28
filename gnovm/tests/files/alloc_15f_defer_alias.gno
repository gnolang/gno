// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

func f(xs ...Big) { println(len(xs)) }

// Regression: deferred-call args were copied into a Go local before the
// Defer was pushed; variant of alloc_15_slicelit_alias.
func main() {
	defer f(a, a, a, a, a, a, a, a, a, a)
	println("deferred")
}

// Error:
// allocation limit exceeded
