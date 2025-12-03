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
// main/cmp_struct_b.gno:15:10-18: invalid operation: fa<VPBlock(1,0)> == bb<VPBlock(1,1)> (mismatched types main.foo and main.bar)

// TypeCheckError:
// main/cmp_struct_b.gno:15:16: invalid operation: fa == bb (mismatched types foo and bar)
