package main

import "filetests/extern/copytest"

func main() {
	copy(copytest.Array[:], []int{1})

	println("OK")
}

// Output:
// OK
