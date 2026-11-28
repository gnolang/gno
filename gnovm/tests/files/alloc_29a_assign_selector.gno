// MAXALLOC: 1000000
package main

type Big [300000]byte

type S struct{ F any }

var a Big

func s() *S { return &S{} }

// Regression: variant of alloc_29_assign_index with a field destination.
func main() {
	s().F, s().F, s().F, s().F = a, a, a, a
	println("survived")
}

// Error:
// allocation limit exceeded
