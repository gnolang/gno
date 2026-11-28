// MAXALLOC: 5000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// StructValue.Copy fills a fresh field buffer one field copy at a time. Each
// field is a 300KB array, so a collection triggered by a later field's copy
// must still count the copies already in the buffer. The cap sits above what
// the program holds with the buffer left out of the count, and below what it
// holds with the buffer counted.
type Big [300000]byte

type S struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9 Big
}

func main() {
	a := new(S)
	b := *a
	println(len(b.F0))
}

// Error:
// allocation limit exceeded
