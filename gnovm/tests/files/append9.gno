package main

import "filetests/extern/copytest"

func main() {
	_ = append(copytest.Slice[:0], 1)
}

// Error:
// cannot modify external-realm or non-realm object
