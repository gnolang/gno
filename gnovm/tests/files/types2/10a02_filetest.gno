package main

func main() {
	x := 11
	y := 1.0 << x
	println(y)
}

// Error:
// main/files/types2/10a02_filetest.gno:5: operator << not defined on: BigdecKind
