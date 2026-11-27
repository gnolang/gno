// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

func t() map[int]Big { return map[int]Big{} }

// Regression: doOpAssign popped its Lhs operands before assigning, so a map
// reached only through them, and every copy stored in it, was invisible to
// GC: four 300KB copies were built under a 1MB cap.
func main() {
	t()[0], t()[1], t()[2], t()[3] = a, a, a, a
	println("survived")
}

// Error:
// allocation limit exceeded
