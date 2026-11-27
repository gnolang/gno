package main

func main() {
	const a int16 = 20000
	println(a * 2)
}

// Error:
// main/const_ovf_mul_int16.gno:5:10-15: a * 2 (constant 40000 of type int16) overflows int16

// TypeCheckError:
// main/const_ovf_mul_int16.gno:5:10: a * 2 (constant 40000 of type int16) overflows int16
