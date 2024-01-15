package main

func main() {
	x := 11
	y := uint64(1.0 << x)
	println(y)
}

// Error:
// main/files/types/27a_filetest.gno:5: operator << not defined on: BigdecKind
