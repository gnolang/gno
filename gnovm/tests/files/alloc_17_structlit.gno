// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// doOpStructLit's ordered branch fills a bare make([]TypedValue, nf) with
// PopCopyValues. Same shape as the slice literal: the fields all alias one
// `a`, so the operand-stack walk charges 300KB while ten fresh 300KB copies
// land in a buffer no root can see.
type Big [300000]byte

type S struct {
	F0 Big
	F1 Big
	F2 Big
	F3 Big
	F4 Big
	F5 Big
	F6 Big
	F7 Big
	F8 Big
	F9 Big
}

var a Big

func main() {
	s := S{a, a, a, a, a, a, a, a, a, a}
	println(len(s.F0))
}

// Error:
// allocation limit exceeded
