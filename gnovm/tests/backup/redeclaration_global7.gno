package main

import (
	"filetests/extern/bar" // pkg name is actually quux
	"filetests/extern/baz" // pkg name is also quux
)

func main() {
	println("Hello", quux.Quux())
}

// Error:
// ../_test/redeclaration_global7.gno:5:2: quux/redeclaration_global7.gno redeclared as imported package name
