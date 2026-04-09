package main

import "fmt"

func main() {
	t := make([]int)
	fmt.Println(t)
}

// TypeCheckError:
// main/varg_8.gno:6:7: invalid operation: make([]int) expects 2 or 3 arguments; found 1

// Error:
// main/varg_8.gno:6:7-18: invalid operation: make([]int) expects 2 or 3 arguments; found 1
