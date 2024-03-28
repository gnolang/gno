package main

import "fmt"

func main() {
	// Define a function that returns a closure
	var recursiveFunc func(int) int
	recursiveFunc = func(num int) int {
		if num <= 0 {
			return 1
		}
		// Closure calling itself recursively
		return num * recursiveFunc(num-1)
	}

	// Use the recursive closure
	result := recursiveFunc(5)
	fmt.Println("Factorial of 5 is:", result) // Output: 120
}

// Output:
// Factorial of 5 is: 120
