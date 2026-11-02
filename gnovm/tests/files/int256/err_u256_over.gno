package main

func main() {
	var x uint256 = 1 << 256
	println(x)
}

// Error:
// main/err_u256_over.gno:4:6-26: bigint overflows target kind

// TypeCheckError:
// main/err_u256_over.gno:4:18: cannot use 1 << 256 (untyped int constant 115792089237316195423570985008687907853269984665640564039457584007913129639936) as uint256 value in variable declaration (overflows)
