package main

type Integer *int

func (f foo) do(i Integer) {
	println(i)
}

type foo struct{}

func main() {
	a := 1

	f := foo{}
	f.do(&a)
}

// Output:
// &(1 int)
