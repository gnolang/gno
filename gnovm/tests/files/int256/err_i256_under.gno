package main

func main() {
	var x int256 = -1<<255 - 1
	println(x)
}

// Error:
// main/err_i256_under.gno:4:6-28: bigint underflows target kind
