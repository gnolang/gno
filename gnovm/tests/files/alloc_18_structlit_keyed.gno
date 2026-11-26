// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// doOpStructLit's keyed branch copies each field named into a buffer that is
// a Go local until the struct is pushed, so only its anchor counts the copies
// already made when a later copy triggers a collection. Selecting one field
// copies only that field, so no other copy of the struct trips the cap first.
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
	x := S{F0: a, F1: a, F2: a, F3: a, F4: a, F5: a, F6: a, F7: a, F8: a, F9: a}.F9
	println(len(x))
}

// Error:
// allocation limit exceeded
