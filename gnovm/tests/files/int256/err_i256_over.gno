package main

func main() {
	var x int256 = 1 << 255
	println(x)
}

// Error:
// main/err_i256_over.gno:4:6-25: bigint overflows target kind
