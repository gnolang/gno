package main

import "fmt"

func main() {
	v := 42
	p := &v
	fmt.Println(p) // Output: memory address of v
}

// Output:
// 0x1400003b0e0
