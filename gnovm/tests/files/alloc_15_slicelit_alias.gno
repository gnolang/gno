// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// doOpSliceLit allocates the backing array with NewListArray and fills it
// with PopCopyValues, so the array is a Go local while every element copy
// lands in it. Visiting the operand stack does not help here: all 100 slots
// alias the same `a`, and vis dedups Objects, so the walk charges one 300KB
// element however many copies get written. GC therefore freed the tally on
// every element and re-granted the cap, building 30MB under a 1MB cap.
//
// Allocator.anchors keeps the buffer visible for the whole fill.
type Big [300000]byte

var a Big

func main() {
	s := []Big{
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
		a, a, a, a, a, a, a, a, a, a,
	}
	println(len(s) * 300000)
}

// Error:
// allocation limit exceeded
