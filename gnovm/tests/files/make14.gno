package main

import "fmt"

func main() {
	a := make([]int, 2147483647.1)
	fmt.Println(len(a))
}

// Error:
// main/files/make14.gno:6:7: invalid length argument for built-in make: 2.1474836471e+09 (untyped float constant) truncated to int

// TypeCheckError:
// main/files/make14.gno:6:19: 2147483647.1 (untyped float constant 2.14748e+09) truncated to int