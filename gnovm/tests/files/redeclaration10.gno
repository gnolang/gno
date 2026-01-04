package main

import _ "filetests/extern/redeclaration3"

func main() {
	println("should not happen")
}

// Error:
// running package "filetests/extern/redeclaration3": redeclarations for identifiers: "a<VPInvalid(0)>.method"
