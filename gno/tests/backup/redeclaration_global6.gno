package main

import (
	"time"
)

func time() string {
	return "hello"
}

func main() {
	t := time()
	println(t)
}

// Error:
// ../_test/redeclaration_global6.gno:7:1: time redeclared in this block
