package main

import "fmt"

func main() {
	a := 1      // a is inferred as int
	b := a << 3 // b is also int
	fmt.Printf("%T, %T, %d, %d\n", a, b, a, b)
}

// Output:
// int, int, 1, 8
