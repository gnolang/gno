package main

import "fmt"

func main() {
	x := "hello"
	y := +x
	fmt.Println(y)
}

// Error:
// main/unary_a0c.gno:7:7-9: operator + not defined on: StringKind

// TypeCheckError:
// main/unary_a0c.gno:7:8: invalid operation: operator + not defined on x (variable of type string)
