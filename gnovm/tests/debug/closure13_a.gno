package main

import "fmt"

func main() {
	var recursiveFunc func(int) int
	var recursiveFunc2 func(int) int

	recursiveFunc = func(num int) int {
		recursiveFunc2 = recursiveFunc

		if num <= 0 {
			return 1
		}

		return num * recursiveFunc2(num-1)
	}

	result := recursiveFunc(5)
	fmt.Println("Factorial of 5 is:", result) // Output: 120
}

// Output:
// Factorial of 5 is: 120
