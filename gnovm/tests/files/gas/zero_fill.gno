// MAXALLOC: 500000000
package main

// Builds zero values element by element: new, var, make by length and by
// capacity, and the fields and elements a literal leaves out. Each element or
// field written is charged once.
type A [1000]int

type S struct {
	X, Y int
	Z    A
}

func main() {
	p := new(A)
	var s S
	l := make([]int, 1000)
	c := make([]int, 0, 1000)
	x := S{}
	y := S{X: 1}
	z := [3]int{1}
	w := [3]byte{1}
	_, _, _, _, _, _, _, _ = p, s, l, c, x, y, z, w
}

// Gas:
// 442325
