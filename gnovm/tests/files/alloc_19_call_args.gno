// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// popCopyArgs copies every argument into a bare make([]TypedValue, ...) that
// nothing references until doOpCall assigns it into the call block. Ten
// aliased 300KB arguments therefore built 3MB under a 1MB cap.
type Big [300000]byte

var a Big

func g(x0, x1, x2, x3, x4, x5, x6, x7, x8, x9 Big) int { return len(x0) }

func main() {
	println(g(a, a, a, a, a, a, a, a, a, a))
}

// Error:
// allocation limit exceeded
