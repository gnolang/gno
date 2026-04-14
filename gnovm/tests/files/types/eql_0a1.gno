package main

// both typed(different) const
func main() {
	println(int(1) != int8(1))
}

// Error:
// main/eql_0a1.gno:5:10-27: invalid operation: (mismatched types int and int8)

// TypeCheckError:
// main/eql_0a1.gno:5:20: invalid operation: int(1) != int8(1) (mismatched types int and int8)
