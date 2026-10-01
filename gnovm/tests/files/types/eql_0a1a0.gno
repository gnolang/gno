package main

func main() {
	a := uint(1)
	println(uint64(1) == a)
}

// Error:
// main/eql_0a1a0.gno:5:10-24: invalid operation: (mismatched types uint64 and uint)

// TypeCheckError:
// main/eql_0a1a0.gno:5:23: invalid operation: uint64(1) == a (mismatched types uint64 and uint)
