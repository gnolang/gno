package main

func main() {
	a := 1
	println(a / 0)
}

// Error:
// main/files/types/4f0_filetest.gno:5: invalid operation: division by zero
