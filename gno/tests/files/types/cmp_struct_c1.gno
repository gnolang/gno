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

	println(bb == fa)
}

// Error:
// main/files/types/cmp_struct_c1.gno:15:10: cannot use main.bar as main.foo without explicit conversion
