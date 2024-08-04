package main

import "fmt"

func main() {
	a := [2]int{1, 2}
	b := [2]int{1, 2}

	fmt.Println("a == b:", a == b)
}

// Output:
// a == b: true
