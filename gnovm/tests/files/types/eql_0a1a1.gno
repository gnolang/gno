package main

func main() {
	a := uint(1)
	println(a == uint64(1))
}

// Error:
// main/files/types/eql_0a1a1.gno:5:10: cannot use uint as uint64

// TypeCheckError:
// main/files/types/eql_0a1a1.gno:5:15: invalid operation: a == uint64(1) (mismatched types uint and uint64)
