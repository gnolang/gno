package main

import "filetests/extern/copytest"

func main() {
	_ = append(copytest.Array1[:0], 1)
}

// Error:
// cannot modify external-realm or non-realm object
