package main

import "filetests/extern/copytest"

func main() {
	copy(copytest.Slice, []int{1})

	println("OK")
}

// Error:
// cannot modify external-realm or non-realm object
