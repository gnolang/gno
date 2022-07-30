package main

import "fmt"

func main() {
	a, b, c := fmt.Println("test")
	println(a, b, c)
}

// Error:
// main/files2/assign12.gno:6: assignment mismatch: 3 variables but fmt<VPBlock(2,0)>.Println returns 2 values