package main

func main() {
	println(int(0) && true) // left typed, right untyped
}

// Error:
// main/files/types/30a3_LAND_filetest.gno:4: operator && not defined on: int
