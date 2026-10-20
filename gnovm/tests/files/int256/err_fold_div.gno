package main

func main() {
	println(int256(-1<<255) / -1)
}

// Error:
// main/err_fold_div.gno:4:10-30: constant 57896044618658097711785492504343953926634992332820282019728792003956564819968 overflows Int256Kind
