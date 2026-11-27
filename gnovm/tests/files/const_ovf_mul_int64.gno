package main

func main() {
	const e int64 = 1 << 62
	println(e * 4)
}

// Error:
// main/const_ovf_mul_int64.gno:5:10-15: e * 4 (constant 18446744073709551616 of type int64) overflows int64

// TypeCheckError:
// main/const_ovf_mul_int64.gno:5:10: e * 4 (constant 18446744073709551616 of type int64) overflows int64
