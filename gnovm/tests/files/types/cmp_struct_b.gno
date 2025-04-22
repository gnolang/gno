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
// main/files/types/cmp_struct_b.gno:15:10: cannot use main.foo as main.bar without explicit conversion

// TypeCheckError:
// main/files/types/cmp_struct_b.gno:15:16: invalid operation: fa == bb (mismatched types foo and bar)
