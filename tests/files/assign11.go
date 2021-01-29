package main

import "fmt"

func main() {
	_, _, _ = fmt.Println("test")
}

// Error:
// assignment mismatch: 3 variables but fmt@@0.Println returns 2 values
