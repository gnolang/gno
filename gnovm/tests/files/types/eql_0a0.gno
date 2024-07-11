package main

// both typed(different) const
func main() {
	println(int(1) == int8(1))
}

// Error:
// main/files/types/eql_0a0.gno:5:10: cannot use int as int8
