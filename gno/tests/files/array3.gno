package main

import "fmt"

func main() {
	a := [3]int{2: 2, 1}
	fmt.Println(a)
}

// Error:
// runtime error: index out of range [3] with length 3
