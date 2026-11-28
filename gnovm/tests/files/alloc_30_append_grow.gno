// MAXALLOC: 1300000
package main

// Regression: append that outgrows a non-nil slice copies both operands into
// a new list, and must keep that list counted while it copies, as append into
// a nil slice does (alloc_28_append_copy.gno).
func main() {
	src := make([][1000]int, 20)
	dst := append([][1000]int{}, src...)
	println(len(src), len(dst))
}

// Error:
// allocation limit exceeded
