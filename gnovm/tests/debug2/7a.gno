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
