package main

import "filetests/extern/copytest"

func main() {
	copy(copytest.Array1[:0], []int{1})

	println("OK")
}

// Output:
// OK
