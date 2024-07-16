package main

type (
	nat   []int
	nat64 []int64
)

func main() {
	var nlist = nat{0}
	var n64list = nat64{1}
	a := append(n64list, nlist...)
	println(a)
}

// Error:
// main/files/assign_unnamed_type/append/named_unnamed_type2_filetest.gno:11:7: cannot use int as int64
