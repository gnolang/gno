package main

type E struct{ X, Y int }

type T struct {
	a struct{ x, y int }
	b int
	E
	c string
}

func main() {
	t := T{b: 1, E: E{X: 2, Y: 3}, c: "c"}
	t.a.x = 4
	println(t.b, t.a.x, t.X, t.Y, t.c)

	p := &T{b: 5, c: "d"}
	p.a.y = 6
	p.X = 7
	println(p.b, p.a.y, p.X, p.c)
}

// Output:
// 1 4 2 3 c
// 5 6 7 d
