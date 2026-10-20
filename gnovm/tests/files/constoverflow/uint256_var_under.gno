package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint256 = -1
	println(x)
}

// Error:
// main/uint256_var_under.gno:7:6-20: bigint underflows target kind
