package main

type L []int

func (l L) Sum() (s int) {
	for _, v := range l {
		s += v
	}
	return
}

type A [3]int

func main() {
	// A nil declared slice, sliced, is a nil L and equals nil, as in Go.
	var empty L
	println(empty[:] == nil)

	// A sliced declared slice keeps the methods of its type.
	l := L{1, 2, 3}
	println(l[1:].Sum())
	println(l[0:2:2].Sum())

	// A sliced declared array, directly or through a pointer, gives an
	// unnamed slice.
	a := A{4, 5, 6}
	var s []int = a[:]
	var p []int = (&a)[1:]
	println(len(s), len(p))
}

// Output:
// true
// 5
// 3
// 3 2
