package main

import "fmt"

func f() (int, int) {
	return 1, 2
}

func main() {
	a, b, c := 1, f(), 3
}

// Error:
// main/files/assign32.gno:10:2: evalStaticTypeOf() only supports *CallExpr with 1 result, got (int,int)
