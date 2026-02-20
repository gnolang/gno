package main

import "fmt"

func main() {
	a := 1/0.5
	println(a)	
	fmt.Printf("%T\n", a)
}

// Output:
// 2
// float64