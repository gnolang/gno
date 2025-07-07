package main

import _ "filetests/extern/redeclaration1"

func main() {
	println("should not happen")
}

// Error:
// running package "filetests/extern/redeclaration1": redeclarations for identifiers: "a"
