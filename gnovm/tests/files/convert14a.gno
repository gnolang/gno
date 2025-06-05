package main

import "fmt"

func main() {
	// b := []rune("hello")
	b := []rune(1)
	fmt.Println(b)
}

// Error:
// main/convert9a.gno:7:7-16: cannot convert (const (1 <untyped> bigint)) to type []int32

// TypeCheckError:
// main/convert9a.gno:7:14: cannot convert 1 (untyped int constant) to type []rune
