package main

import "fmt"

func main() {
	x := 10
	p := &x
	fmt.Println(*p) // Output: 10
}

// Output:
// 10
