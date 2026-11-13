package main

type T struct{ G int }

func (t *T) SetG(v int) { t.G = v }

type S struct {
	F int
	T T
}

func main() {
	m := map[int]S{1: {F: 1}}
	m[1].F = 9
	println(m[1].F)
}

// Error:
// main/addressable_12i_err.gno:14:2-12: cannot assign to m<VPBlock(1,0)>[(const (1 int))].F

// TypeCheckError:
// main/addressable_12i_err.gno:14:2: cannot assign to struct field m[1].F in map
