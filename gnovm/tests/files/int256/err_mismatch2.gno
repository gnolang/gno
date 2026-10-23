package main

func main() {
	var a int256 = 1
	var b uint256 = 2
	println(a == b)
}

// Error:
// main/err_mismatch2.gno:6:10-16: invalid operation: (mismatched types int256 and uint256)

// TypeCheckError:
// main/err_mismatch2.gno:6:15: invalid operation: a == b (mismatched types int256 and uint256)
