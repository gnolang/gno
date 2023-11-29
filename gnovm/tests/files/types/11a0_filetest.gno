package main

// both typed(different) const
func main() {
	println(int(1) && int(1))
}

// Error:
// main/files/types/11a0_filetest.gno:5: operator && not defined on: int