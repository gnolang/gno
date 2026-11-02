package main

func main() {
	var x int256 = -1<<255 - 1
	println(x)
}

// Error:
// main/err_i256_under.gno:4:6-28: bigint underflows target kind

// TypeCheckError:
// main/err_i256_under.gno:4:17: cannot use -1 << 255 - 1 (untyped int constant -57896044618658097711785492504343953926634992332820282019728792003956564819969) as int256 value in variable declaration (overflows)
