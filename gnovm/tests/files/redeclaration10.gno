package main

import _ "github.com/gnolang/gno/_test/redeclaration3"

func main() {
	println("should not happen")
}

// XXX show what was redeclared.

// Error:
// running package "github.com/gnolang/gno/_test/redeclaration3": duplicate declarations not allowed
