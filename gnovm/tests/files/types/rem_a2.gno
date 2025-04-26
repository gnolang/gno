package main

func main() {
	println(1 % 0)
}

// Error:
// main/files/types/rem_a2.gno:4:10: invalid operation: division by zero

// TypeCheckError:
// main/files/types/rem_a2.gno:4:14: invalid operation: division by zero
