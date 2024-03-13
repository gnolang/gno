package main

import "fmt"

func main() {

	var fns []func(int) int
	var recursiveFunc func(int) int

	for i := 0; i < 6; i++ {
		recursiveFunc = func(num int) int {
			if num <= 0 {
				return 1
			}
			return num * recursiveFunc(num-1)
		}
		fns = append(fns, recursiveFunc)
	}

	for i, r := range fns {
		result := r(i)
		fmt.Printf("Factorial of %d is: %d \n", i, result)
	}
}

// Output:
// Factorial of 0 is: 1
// Factorial of 1 is: 1
// Factorial of 2 is: 2
// Factorial of 3 is: 6
// Factorial of 4 is: 24
// Factorial of 5 is: 120
