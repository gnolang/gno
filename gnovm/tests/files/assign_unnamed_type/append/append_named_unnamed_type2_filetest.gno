package main

type (
	nat   []int
	nat64 []int64
)

func main() {
	nlist := nat{0}
	n64list := nat64{1}
	a := append(n64list, nlist...)
	println(a)
}

// Error:
// main/files/assign_unnamed_type/append/append_named_unnamed_type2_filetest.gno:11: cannot use int as int64
