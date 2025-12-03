package main

func main() {
	var a int = 1
	var b float32 = 1.0
	println(a == b) // both typed
}

// Error:
// main/eql_0a1g.gno:6:10-16: invalid operation: a<VPBlock(1,0)> == b<VPBlock(1,1)> (mismatched types int and float32)

// TypeCheckError:
// main/eql_0a1g.gno:6:15: invalid operation: a == b (mismatched types int and float32)
