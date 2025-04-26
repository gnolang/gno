package main

import "fmt"

func f() (a, b int) {
	return 1, 2
}

func main() {
	x, y, z := 1, f()
	fmt.Println(x, y, z)
}

// Error:
// main/files/assign25c.gno:10:2: assignment mismatch: 3 variable(s) but 2 value(s)

// TypeCheckError:
// main/files/assign25c.gno:10:13: assignment mismatch: 3 variables but 2 values
