package main

import "fmt"

func main() {
	a := 5 // infer type int
	var b int32 = 10
	c := b + a<<2

	fmt.Printf("%T, %d\n", c, c)
}

// Error:
// main/shift_d54.gno:8:7-15: invalid operation: b<VPBlock(1,1)> + a<VPBlock(1,0)> << (const (2 uint)) (mismatched types int32 and int)

// TypeCheckError:
// main/shift_d54.gno:8:7: invalid operation: b + a << 2 (mismatched types int32 and int)
