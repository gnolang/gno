package main

type MyInt int8

func main() {
	const c MyInt = 100
	println(c * 2)
}

// Error:
// main/const_ovf_mul_named.gno:7:10-15: c * 2 (constant 200 of int8 type MyInt) overflows int8

// TypeCheckError:
// main/const_ovf_mul_named.gno:7:10: c * 2 (constant 200 of int8 type MyInt) overflows int8
