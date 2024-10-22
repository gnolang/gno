package main

import "fmt"

func main() {
	_, _, _ = fmt.Println("test")
}

// Error:
// main/files/assign11.gno:6:2: assignment mismatch: 3 variables but fmt<VPBlock(2,0)>.Println returns 2 values
