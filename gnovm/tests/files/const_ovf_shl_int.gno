package main

func main() {
	const a int = 1
	println(a << 63)
}

// Error:
// main/const_ovf_shl_int.gno:5:10-17: a << 63 (constant 9223372036854775808 of type int) overflows int

// TypeCheckError:
// main/const_ovf_shl_int.gno:5:10: a << 63 (constant 9223372036854775808 of type int) overflows int
