package main

type foo struct {
	a int
}

type bar struct {
	b int
}

func main() {
	fa := foo{}
	bb := bar{}

	println(fa == bb)
}

// Error:
// main/cmp_struct_b.gno:15:10-18: invalid operation: (mismatched types main.foo and main.bar)

// TypeCheckError:
// main/cmp_struct_b.gno:15:16: invalid operation: fa == bb (mismatched types foo and bar)
