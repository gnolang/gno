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
// main/named_unnamed_type2.gno:11:7-32: cannot use int as int64

// TypeCheckError:
// main/named_unnamed_type2.gno:11:23: cannot use nlist (variable of slice type nat) as []int64 value in argument to append
