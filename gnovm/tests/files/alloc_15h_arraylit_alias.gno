// MAXALLOC: 1000000
package main

type Big [300000]byte

var a Big

// NOT a regression test for doOpArrayLit's anchor, despite the shape: it
// passes with that anchor removed, because `s :=` then copies the finished
// literal under ArrayValue.Copy's own anchor, which trips the 1MB cap either
// way. Kept as a cap-enforcement case for the array literal;
// alloc_25_arraylit_index.gno indexes the literal instead, so only the
// literal's anchor counts its elements.
func main() {
	s := [10]Big{a, a, a, a, a, a, a, a, a, a}
	println(len(s))
}

// Error:
// allocation limit exceeded
