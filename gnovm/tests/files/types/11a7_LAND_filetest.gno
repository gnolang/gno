package main

func main() {
	println(int(0) && true) // left typed, right untyped
}

// Error:
// main/files/types/11a7_LAND_filetest.gno:4: operator && not defined on: int
