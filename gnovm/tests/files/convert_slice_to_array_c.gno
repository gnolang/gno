package main

type Point struct {
	X, Y int
}

func main() {
	s := []Point{{1, 2}, {3, 4}, {5, 6}}
	a := [2]Point(s)
	a[0].X = 99
	println(a[0].X, a[0].Y, a[1].X, a[1].Y)
	println(s[0].X, s[0].Y)
}

// Output:
// 99 2 3 4
// 1 2
