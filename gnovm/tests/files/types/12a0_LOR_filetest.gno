package main

// both typed(different) const
func main() {
	println(int(1) || int(1))
}

// Error:
// main/files/types/12a0_LOR_filetest.gno:5: operator || not defined on: IntKind
