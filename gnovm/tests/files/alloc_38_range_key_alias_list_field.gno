// MAXALLOC: 1100000
package main

// Regression: a range's list element stays counted while its key writes
// inside it.
//
// `for s[0].N, v = range s` copies the element before writing the key into
// it, so the value written to v is the element as read. The copy is a Go local
// until written, while the element itself stays live in s, so only the copy's
// anchor counts it when writing it to v copies it again. alloc_35 trips at the
// first copy instead, before the anchor holds anything.
const N = 300000

type S struct {
	N int
	A [N]byte
}

func main() {
	s := []S{{}}
	var v S
	for s[0].N, v = range s {
	}
	println(len(v.A))
}

// Error:
// allocation limit exceeded
