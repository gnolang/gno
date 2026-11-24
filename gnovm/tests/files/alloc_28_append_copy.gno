// MAXALLOC: 1300000
package main

// Regression: append into a new list must keep that list counted while it
// copies the elements in.
//
// Each element of src is a [1000]int, so each copy allocates, and a collection
// triggered by a later copy must still count the copies already made. Left
// out of the count, append built a second 20-element copy under a cap that
// refuses make([][1000]int, 20) followed by copy.
func main() {
	src := make([][1000]int, 20)
	dst := append([][1000]int(nil), src...)
	println(len(src), len(dst))
}

// Error:
// allocation limit exceeded
