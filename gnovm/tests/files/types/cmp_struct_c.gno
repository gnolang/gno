package main

type foo struct {
	a int
}

type bar struct {
	b []int
}

func main() {
	fa := foo{}
	bb := bar{}

	println(fa == bb)
}

// Error:
// main/files/types/cmp_struct_c.gno:15:10: main.bar is not comparable

// TypeCheckError:
// main/files/types/cmp_struct_c.gno:15:16: invalid operation: fa == bb (mismatched types foo and bar)
