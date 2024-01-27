package main

import _ "github.com/gnolang/gno/_test/redeclaration1"

func main() {
	println("should not happen")
}

// Error:
// running package "github.com/gnolang/gno/_test/redeclaration1": invalid redeclarations for identifiers: "a"
