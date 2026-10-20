package main

func main() {
	println(uint256(1) << 256)
}

// Error:
// main/err_fold_shift.gno:3:1-5:2: constant 115792089237316195423570985008687907853269984665640564039457584007913129639936 overflows Uint256Kind
