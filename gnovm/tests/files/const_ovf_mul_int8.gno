package main

func main() {
	const c int8 = 100
	println(c * 2)
}

// Error:
// main/const_ovf_mul_int8.gno:5:10-15: c * 2 (constant 200 of type int8) overflows int8

// TypeCheckError:
// main/const_ovf_mul_int8.gno:5:10: c * 2 (constant 200 of type int8) overflows int8
