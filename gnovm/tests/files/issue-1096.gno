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
	println(x)
	x.manip()
	println(x)

	println("-----")

	y := Y{}
	y[1] = 888
	println(y)
	y.manip()
	println(y)

	println("-----")

	x = X{}
	println(x)
	x.Array[1] = 888
	println(x)
	manip(x)
	println(x)
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
// (struct{(array[(0 int),(888 int),(0 int),(0 int),(0 int),(0 int),(0 int),(0 int)] [8]int),(false bool)} main.X)
// (struct{(array[(0 int),(888 int),(0 int),(0 int),(0 int),(0 int),(0 int),(0 int)] [8]int),(false bool)} main.X)
// -----
// (array[(0 int),(888 int),(0 int),(0 int),(0 int),(0 int),(0 int),(0 int)] main.Y)
// (array[(0 int),(888 int),(0 int),(0 int),(0 int),(0 int),(0 int),(0 int)] main.Y)
// -----
// (struct{(array[(0 int),(0 int),(0 int),(0 int),(0 int),(0 int),(0 int),(0 int)] [8]int),(false bool)} main.X)
// (struct{(array[(0 int),(888 int),(0 int),(0 int),(0 int),(0 int),(0 int),(0 int)] [8]int),(false bool)} main.X)
// (struct{(array[(0 int),(888 int),(0 int),(0 int),(0 int),(0 int),(0 int),(0 int)] [8]int),(false bool)} main.X)
