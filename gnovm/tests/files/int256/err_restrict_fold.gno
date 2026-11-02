package main

// The restriction applies to typed int256 constant arithmetic too.
const c int256 = 1 << 62

func main() {
	println(c * 4)
}

// Output:
// 18446744073709551616

// TypeCheckError:
// main/err_restrict_fold.gno:7:10: c * 4 (constant 18446744073709551616 of int type int256) overflows int
