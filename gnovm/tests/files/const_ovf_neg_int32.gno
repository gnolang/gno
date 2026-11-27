package main

func main() {
	const g int32 = -2147483648
	println(-g)
}

// Error:
// main/const_ovf_neg_int32.gno:5:10-12: -g (constant 2147483648 of type int32) overflows int32

// TypeCheckError:
// main/const_ovf_neg_int32.gno:5:10: -g (constant 2147483648 of type int32) overflows int32
