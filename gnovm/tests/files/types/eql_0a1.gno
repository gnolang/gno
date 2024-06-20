package main

// both typed(different) const
func main() {
	println(int(1) != int8(1))
}

// Error:
// main/files/types/eql_0a1.gno:5: cannot use int as int8
