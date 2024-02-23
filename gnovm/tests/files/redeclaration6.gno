package main

import _ "github.com/gnolang/gno/_test/redeclaration1"

func main() {
	println("should not happen")
}

// XXX show what was redeclared.

// Error:
// running package "github.com/gnolang/gno/_test/redeclaration1": duplicate declarations not allowed
