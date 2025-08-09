package main

import (
	"fmt"
)

func main() {
	var nilSlice []int // nil slice

	// Convert nil slice to [0]int (zero-length array)
	zeroArray := [0]int(nilSlice)

	fmt.Println(zeroArray == nil) // type mismatch
}

// Error:
// main/convert11c.gno:13:14-30: invalid operation: zeroArray<VPBlock(1,1)> == (const (undefined)) (mismatched types [0]int and untyped nil)

// TypeCheckError:
// main/convert11c.gno:13:27: invalid operation: zeroArray == nil (mismatched types [0]int and untyped nil)
