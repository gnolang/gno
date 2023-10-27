package main

import _ "gno.land/p/demo/tests/invalid/redeclaration3"

func main() {
	println("should not happen")
}

// XXX show what was redeclared.

// Error:
// running package "gno.land/p/demo/tests/invalid/redeclaration3": duplicate declarations not allowed
