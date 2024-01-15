package main

// both typed(different) const
func main() {
	println(true || 1)
}

// Error:
// main/files/types/12a2_filetest.gno:5: operator || not defined on: BigintKind
