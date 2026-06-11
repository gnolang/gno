package main

func main() {
	_ = make(map[int]int, 1, 2)
}

// TypeCheckError:
// main/varg_10.gno:4:6: invalid operation: make(map[int]int, 1, 2) expects 1 or 2 arguments; found 3

// Error:
// main/varg_10.gno:4:6-29: invalid operation: make(map[int]int) expects 1 or 2 arguments; found 3
