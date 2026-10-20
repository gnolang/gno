package main

func main() {
	println(-uint256(1))
}

// Output:
// 115792089237316195423570985008687907853269984665640564039457584007913129639935

// TypeCheckError:
// main/err_neg_uint.gno:4:10: -uint256(1) (constant -1 of uintptr type uint256) overflows uintptr
