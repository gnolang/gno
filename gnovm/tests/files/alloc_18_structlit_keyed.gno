// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// doOpStructLit's keyed branch fills the field buffer twice — once with
// defaultStructFields' zero values, then with the keyed copies — and both
// fills ran against a buffer that was invisible to GC.
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
	s := S{F0: a, F1: a, F2: a, F3: a, F4: a, F5: a, F6: a, F7: a, F8: a, F9: a}
	println(len(s.F0))
}

// Error:
// allocation limit exceeded
