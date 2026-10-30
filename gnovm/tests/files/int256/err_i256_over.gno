package main

func main() {
	var x int256 = 1 << 255
	println(x)
}

// Error:
// main/err_i256_over.gno:4:6-25: bigint overflows target kind

// TypeCheckError:
// main/err_i256_over.gno:4:17: cannot use 1 << 255 (untyped int constant 57896044618658097711785492504343953926634992332820282019728792003956564819968) as int256 value in variable declaration (overflows)
