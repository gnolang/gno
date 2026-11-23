// MAXALLOC: 1000000
package main

// Regression: make([]T, l, c) fills every element up to its capacity, and must
// keep that list counted as make([]T, n) does. See alloc_26_make_len.gno.
func main() {
	s := make([][1000]int, 0, 100)
	println(len(s), cap(s))
}

// Error:
// allocation limit exceeded
