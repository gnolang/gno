package main

import "fmt"

func main() {
	x := "hello"
	y := +x
	fmt.Println(y)
}

// Error:
// main/files/types/unary_a0c.gno:7:7: operator + not defined on: StringKind
