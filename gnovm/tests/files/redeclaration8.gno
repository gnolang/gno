package main

import _ "filetests/extern/redeclaration2"

func main() {
	println("should not happen")
}

// Error:
// running package "filetests/extern/redeclaration2": redeclarations for identifiers: "a"
