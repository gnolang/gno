package main

import "fmt"

// recursive closure does not capture
func main() {
	var fns []func(int) int
	var recursiveFunc func(int) int

	for i := 0; i < 3; i++ {
		recursiveFunc = func(num int) int {
			x := i
			println("value of x: ", x)
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
// value of x:  0
// Factorial of 0 is: 1
// value of x:  1
// value of x:  2
// Factorial of 1 is: 1
// value of x:  2
// value of x:  2
// value of x:  2
// Factorial of 2 is: 2
