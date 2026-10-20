package main

func main() {
	println(uint256(0) - 1)
}

// Error:
// main/err_fold_sub.gno:4:10-24: constant -1 overflows Uint256Kind
