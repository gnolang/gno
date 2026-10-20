package main

func main() {
	println(-uint256(1))
}

// Error:
// main/err_neg_uint.gno:4:10-21: constant -1 overflows Uint256Kind
