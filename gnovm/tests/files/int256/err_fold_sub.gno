package main

func main() {
	println(uint256(0) - 1)
}

// Output:
// 115792089237316195423570985008687907853269984665640564039457584007913129639935

// TypeCheckError:
// main/err_fold_sub.gno:4:10: uint256(0) - 1 (constant -1 of uintptr type uint256) overflows uintptr
