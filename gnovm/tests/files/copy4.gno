package main

import "filetests/extern/copytest"

func main() {
	// Appending nil to these is not a problem; if either operand is nil, copy
	// is considered a read-only no-op.
	copy(copytest.Slice, []int(nil))
	copy(copytest.Array[:], []int(nil))
	copy(copytest.Array1[:0], []int(nil))
	copy(copytest.Array1[:], []int(nil))
	copy(copytest.Array1[:0:0], []int(nil))

	println("OK")
}

// Output:
// OK
