package main

import "fmt"

func main() {
	a, b, c := fmt.Println("test")
	println(a, b, c)
}

// Error:
// assignment mismatch: 3 variables but fmt<VPBlock(2,0)>.Println returns 2 values
