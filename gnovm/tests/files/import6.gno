package main

import "filetests/extern/c1"

func main() {
	println(c1.C1)
}

// Error:
// filetests/extern/c2/c2.gno:3:8-29: import cycle detected: "filetests/extern/c1" (through [filetests/extern/c1 filetests/extern/c2])
