package main

import (
	"github.com/gnolang/gno/_test/bar" // pkg name is actually quux
)

func main() {
	println("Hello", bar.Quux()) // bar should not be a known symbol.
}

// Error:
// name bar not declared
