package main

// A conversion of nil to a pointer to an array, named or not, is a constant
// conversion like any other, not a len or cap to fold.

type Hash [32]byte

func (h *Hash) Len() int { return len(h) }

type Lener interface{ Len() int }

var _ Lener = (*Hash)(nil)

func main() {
	p := (*Hash)(nil)
	q := (*[3]int)(nil)
	println(p == nil, q == nil)
	var l Lener = (*Hash)(nil)
	println(l.Len(), len(q))
}

// Output:
// true true
// 32 3
