// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// doOpSliceLit2 (indexed slice literal) fills its backing array twice: the
// keyed copies, then defaultTypedValue for every index left undefined. Both
// fills wrote into a Go local. The default fill is the worse half — the
// index range, not the element count, decides how much it allocates.
type Big [300000]byte

var a Big

func main() {
	s := []Big{0: a, 1: a, 2: a, 3: a, 4: a, 5: a, 6: a, 7: a, 8: a, 9: a}
	println(len(s))
}

// Error:
// allocation limit exceeded
