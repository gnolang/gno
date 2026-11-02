package main

func main() {
	println(uint256(1) << 256)
}

// Output:
// 0

// TypeCheckError:
// main/err_fold_shift.gno:4:10: uint256(1) << 256 (constant 115792089237316195423570985008687907853269984665640564039457584007913129639936 of uintptr type uint256) overflows uintptr
