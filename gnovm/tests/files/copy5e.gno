package main

import "filetests/extern/copytest"

// copy5e — copying into an external-realm, Data-backed []byte exercises the
// copy() byte fast path (uverse.go), which writes raw bytes directly without
// going through Assign2. The readonly-taint check must still guard it, just
// like the non-byte slow path (copy5/copy5c) and append.

func main() {
	copy(copytest.DataByteSlice, []byte{9})

	println("OK")
}

// Error:
// cannot copy to readonly tainted slice
