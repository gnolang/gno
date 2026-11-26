package main

import "filetests/extern/copytest"

func main() {
	_ = append(copytest.Array1[:0], 1)
}

// Error:
// cannot append to readonly tainted slice (call a crossing function of the owning realm)
