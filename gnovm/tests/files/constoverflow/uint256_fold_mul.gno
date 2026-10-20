package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint256 = 1<<256 - 1

func main() {
	println(c * 2)
}

// Error:
// main/uint256_fold_mul.gno:9:10-15: constant 231584178474632390847141970017375815706539969331281128078915168015826259279870 overflows Uint256Kind
