package main

const c uint256 = 1<<256 - 1

func main() {
	const d = c + 1
	println(d)
}

// Error:
// main/err_fold_add.gno:6:12-17: constant 115792089237316195423570985008687907853269984665640564039457584007913129639936 overflows Uint256Kind
