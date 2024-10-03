package main

import (
	"github.com/gnolang/gno/_test/bar" // pkg name is actually quux
	"github.com/gnolang/gno/_test/baz" // pkg name is also quux
)

func main() {
	println("Hello", quux.Quux())
}

// Error:
// ../_test/redeclaration_global7.gno:5:2: quux/redeclaration_global7.gno redeclared as imported package name
