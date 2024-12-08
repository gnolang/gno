package main

import _ "github.com/gnolang/gno/_test/redeclaration3"

func main() {
	println("should not happen")
}

// Error:
// running package "github.com/gnolang/gno/_test/redeclaration3": redeclarations for identifiers: "a<VPUverse(0)>.method"
