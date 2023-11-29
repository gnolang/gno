package main

// both typed(different) const
func main() {
	int(0) += int8(1)
}

// Error:
// main/files/types/13a0_filetest.gno:5: cannot use int8 as int
