package main

func main() {
	const a int8 = -1
	println(a << 8)
}

// Error:
// main/const_ovf_shl_neg_int8.gno:5:10-16: a << 8 (constant -256 of type int8) overflows int8

// TypeCheckError:
// main/const_ovf_shl_neg_int8.gno:5:10: a << 8 (constant -256 of type int8) overflows int8
