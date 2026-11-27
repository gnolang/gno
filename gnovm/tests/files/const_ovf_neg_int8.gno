package main

func main() {
	const a int8 = -128
	println(-a)
}

// Error:
// main/const_ovf_neg_int8.gno:5:10-12: -a (constant 128 of type int8) overflows int8

// TypeCheckError:
// main/const_ovf_neg_int8.gno:5:10: -a (constant 128 of type int8) overflows int8
