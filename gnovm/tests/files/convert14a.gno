package main

import "fmt"

func main() {
	b := []rune(1)
	fmt.Println(b)
}

// Error:
// main/convert14a.gno:6:7-16: cannot convert (const (1 <untyped> bigint)) to type []int32

// TypeCheckError:
// main/convert14a.gno:6:14: cannot convert 1 (untyped int constant) to type []rune
