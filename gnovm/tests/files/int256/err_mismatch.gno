package main

func main() {
	var a int256 = 1
	var b int = 2
	println(a + b)
}

// Error:
// main/err_mismatch.gno:6:10-15: invalid operation: a<VPBlock(1,0)> + b<VPBlock(1,1)> (mismatched types int256 and int)

// TypeCheckError:
// main/err_mismatch.gno:6:10: invalid operation: a + b (mismatched types int256 and int)
