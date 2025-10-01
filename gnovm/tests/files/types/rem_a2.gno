package main

func main() {
	println(1 % 0)
}

// Error:
// main/rem_a2.gno:4:10-15: invalid operation: division by zero

// TypeCheckError:
// main/rem_a2.gno:4:14: invalid operation: division by zero
