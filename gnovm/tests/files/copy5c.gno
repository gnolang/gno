package main

import "filetests/extern/copytest"

func main() {
	copy(copytest.Array1[:], []int{1})

	println("OK")
}

// Error:
// cannot copy to readonly tainted slice
