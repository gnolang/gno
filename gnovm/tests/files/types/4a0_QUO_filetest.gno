package main

// both typed(different) const
func main() {
	println(int(1) / int8(1))
}

// Error:
// main/files/types/4a0_QUO_filetest.gno:5: cannot use int as int8
