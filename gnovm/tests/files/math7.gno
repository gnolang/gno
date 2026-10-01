package main

import (
	"fmt"
)

func main() {
	var a float64 = 0.0
	var b float32 = 0.0
	fmt.Println(1 / a)
	fmt.Println(0 / a)
	fmt.Println(-1 / a)
	fmt.Println(1 / b)
	fmt.Println(0 / b)
	fmt.Println(-1 / b)
}

// Output:
// +Inf
// NaN
// -Inf
// +Inf
// NaN
// -Inf
