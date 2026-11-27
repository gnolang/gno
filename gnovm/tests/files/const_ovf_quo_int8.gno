package main

func main() {
	const a int8 = -128
	println(a / -1)
}

// Error:
// main/const_ovf_quo_int8.gno:5:10-16: a / -1 (constant 128 of type int8) overflows int8

// TypeCheckError:
// main/const_ovf_quo_int8.gno:5:10: a / -1 (constant 128 of type int8) overflows int8
