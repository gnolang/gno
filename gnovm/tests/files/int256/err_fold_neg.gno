package main

const m int256 = -1 << 255

func main() {
	println(-m)
}

// Error:
// main/err_fold_neg.gno:6:10-12: constant 57896044618658097711785492504343953926634992332820282019728792003956564819968 overflows Int256Kind
