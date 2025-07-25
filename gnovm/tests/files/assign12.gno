package main

import "fmt"

func main() {
	a, b, c := fmt.Println("test")
	println(a, b, c)
}

// Error:
// main/assign12.gno:6:2-32: assignment mismatch: 3 variables but fmt<VPBlock(2,0)>.Println returns 2 values

// TypeCheckError:
// main/assign12.gno:6:13: assignment mismatch: 3 variables but fmt.Println returns 2 values
