package main

import _ "github.com/gnolang/gno/_test/redeclaration2"

func main() {
	println("should not happen")
}

// XXX show what was redeclared.

// Error:
// running package "github.com/gnolang/gno/_test/redeclaration2": invalid redeclarations for identifiers: "a"
