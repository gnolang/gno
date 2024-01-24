package main

import "fmt"

// recursive closure does not capture
func main() {

	var fns []func(int) int
	var recursiveFunc func(int) int

	for i := 0; i < 3; i++ {
		recursiveFunc = func(num int) int {
			x := i
			if num <= 0 {
				return 1
			}
			println(x)
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
// 2
// Factorial of 1 is: 1
// 2
// 2
// Factorial of 2 is: 2
