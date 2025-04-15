package main

import "fmt"

type Foo struct{}

func (foo *Foo) blah(x int, y int) (z int, cb1 func()) {
	a := func(b int, c int) (d int, e int, cb2 func()) {
		return 0, 100, func() {
			b++
			c++
			d++
			e++
			x++
			y++
			z++
			fmt.Println(b, c, d, e, x, y, z) // 201 301 1 101 -99 -99 401
		}
	}
	d2, e2, cb2 := a(200, 300)
	fmt.Println(d2, e2) // 0 100
	x, y, z = -100, -100, -100
	return 400, cb2
}

func main() {
	foo := &Foo{}
	z, cb1 := foo.blah(500, 600)
	cb1()
	fmt.Println(z) // 400
}

// Output:
// 0 100
// 201 301 1 101 -99 -99 401
// 400
