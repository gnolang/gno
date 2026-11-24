// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

func p() *any { return new(any) }

// Regression: variant of alloc_29_assign_index with a pointer destination.
func main() {
	*p(), *p(), *p(), *p() = a, a, a, a
	println("survived")
}

// Error:
// allocation limit exceeded
