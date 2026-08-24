package main

func main() {
	_ = make([]int, 1, 2, 3)
}

// TypeCheckError:
// main/varg_9.gno:4:6: invalid operation: make([]int, 1, 2, 3) expects 2 or 3 arguments; found 4

// Error:
// main/varg_9.gno:4:6-26: invalid operation: make([]int) expects 2 or 3 arguments; found 4
