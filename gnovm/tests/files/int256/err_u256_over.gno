package main

func main() {
	var x uint256 = 1 << 256
	println(x)
}

// Error:
// main/err_u256_over.gno:4:6-26: bigint overflows target kind
