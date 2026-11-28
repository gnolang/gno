// MAXALLOC: 1000000
package main

// Regression: make([]T, n) must keep the list it fills counted.
//
// Each element of a [][1000]int is a fresh [1000]int zero value, so a
// collection triggered by a later element must still count the elements
// already built. Left out of the count, the cap was granted again at each
// collection and the 100 elements, well over 1MB, were built under a 1MB cap
// that refuses `var a [100][1000]int`.
func main() {
	s := make([][1000]int, 100)
	println(len(s), s[99][999])
}

// Error:
// allocation limit exceeded
