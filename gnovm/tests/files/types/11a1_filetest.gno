package main

// both typed(different) const
func main() {
	println(1 && 1)
}

// Error:
// main/files/types/11a1_filetest.gno:5: operator && not defined on: BigintKind
