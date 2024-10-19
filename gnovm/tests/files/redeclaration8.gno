package main

import _ "github.com/gnolang/gno/_test/redeclaration2"

func main() {
	println("should not happen")
}

// Error:
// running package "github.com/gnolang/gno/_test/redeclaration2": redeclarations for identifiers: "a"
