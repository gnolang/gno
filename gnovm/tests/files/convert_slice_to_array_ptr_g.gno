// Non-byte struct elements: pointer mutation propagates to slice.
package main

type Point struct {
	X, Y int
}

func main() {
	s := []Point{{1, 2}, {3, 4}, {5, 6}}
	p := (*[2]Point)(s)
	(*p)[0].X = 99
	(*p)[1].Y = 88
	println(s[0].X, s[0].Y, s[1].X, s[1].Y, s[2].X, s[2].Y)
}

// Output:
// 99 2 3 88 5 6
