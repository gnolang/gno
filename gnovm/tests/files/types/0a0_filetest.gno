package main

// both typed(different) const
func main() {
	println(int(1) == int8(1))
}

// Error:
// main/files/types/0a0_filetest.gno:5: cannot use int as int8
