package main

func main() {
	a := 1
	println(a % 0)
}

// Error:
// main/rem_f3.gno:5:10-15: invalid operation: division by zero

// TypeCheckError:
// main/rem_f3.gno:5:14: invalid operation: division by zero
