package main

import (
	"github.com/gnolang/gno/_test/bar" // pkg name is actually quux
	"github.com/gnolang/gno/_test/baz" // pkg name is also quux
)

func main() {
	println("Hello", quux.Quux())
}

// Error:
// ../_test/redeclaration-global7.go:5:2: quux/redeclaration-global7.go redeclared as imported package name
