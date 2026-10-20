package main

func main() {
	var x uint256 = -1
	println(x)
}

// Error:
// main/err_u256_neg.gno:4:6-20: bigint underflows target kind

// TypeCheckError:
// main/err_u256_neg.gno:4:18: cannot use -1 (untyped int constant) as uint256 value in variable declaration (overflows)
