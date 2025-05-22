package main

type foo struct {
	a []int
}

func main() {
	fa := foo{}
	fb := foo{}

	println(fa == fb)
}

// Error:
// main/files/types/cmp_struct_e.gno:11:10: main.foo is not comparable

// TypeCheckError:
// main/files/types/cmp_struct_e.gno:11:10: invalid operation: fa == fb (struct containing []int cannot be compared)
