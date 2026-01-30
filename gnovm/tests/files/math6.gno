package main

import (
	"fmt"
)

func main() {
	var a float64 = 0.0
	var b float32 = 0.0
	fmt.Println(a / a)
	fmt.Println(b / b)
}

// Output:
// NaN
// NaN
