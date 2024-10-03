package main

import "fmt"

type X struct {
	Array [8]int
	Test  bool
}

type Y [8]int

func main() {
	x := X{}
	x.Array[1] = 888
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	x.manip()
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("-----")

	y := Y{}
	y[1] = 888
	println(y[1])
	println(y[2])

	y.manip()
	println(y[1])
	println(y[2])
	println("-----")

	x = X{}
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	x.Array[1] = 888
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	manip(x)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)
}

func (x X) manip() {
	x.Array[2] = 999
	x.Test = true
}

func manip(x X) {
	x.Array[2] = 999
	x.Test = true
}

func (y Y) manip() {
	y[2] = 111
}

// Output:
// 888
// 0
// false
// 888
// 0
// false
// -----
// 888
// 0
// 888
// 0
// -----
// 0
// 0
// false
// 888
// 0
// false
// 888
// 0
// false
