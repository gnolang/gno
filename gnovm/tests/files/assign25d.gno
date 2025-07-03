package main

import "fmt"

func f() (a, b int) {
	return 1, 2
}

func main() {
	x, y, z := f(), 1, 2, 3
	fmt.Println(x, y, z)
}

// Error:
// main/assign25d.gno:10:2-25: assignment mismatch: 3 variable(s) but 4 value(s)

// TypeCheckError:
// main/assign25d.gno:10:13: assignment mismatch: 3 variables but 4 values
