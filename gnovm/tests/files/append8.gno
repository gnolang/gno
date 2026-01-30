package main

import "filetests/extern/copytest"

func main() {
	// These are all allowed, as they create new slices.
	_ = append(copytest.Slice, 1)
	_ = append(copytest.Slice[:0:0], 1)
	_ = append(copytest.Array[:], 1)
	_ = append(copytest.Array[:0], 1) // original array is empty, will create a new backing array.
	_ = append(copytest.Array1[:0:0], 1)

	println("OK")
}

// Output:
// OK
