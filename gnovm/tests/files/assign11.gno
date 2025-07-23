package main

import "fmt"

func main() {
	_, _, _ = fmt.Println("test")
}

// Error:
// main/assign11.gno:6:2-31: assignment mismatch: 3 variables but fmt<VPBlock(2,0)>.Println returns 2 values

// TypeCheckError:
// main/assign11.gno:6:12: assignment mismatch: 3 variables but fmt.Println returns 2 values
