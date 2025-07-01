package main

import "fmt"

func main() {
	a, c := 1, 2, 3
	fmt.Println(a, c)
}

// Error:
// main/assign28.gno:6:2-17: assignment mismatch: 2 variable(s) but 3 value(s)

// TypeCheckError:
// main/assign28.gno:6:10: assignment mismatch: 2 variables but 3 values
