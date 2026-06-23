package main

import "filetests/extern/copytest"

// copy5f — copying into an external-realm, Data-backed *named* byte slice
// (type NamedBytes []byte) still resolves through baseOf to the copy() byte
// fast path (uverse.go), so the readonly-taint check must guard it just like
// the unnamed case (copy5e).

func main() {
	copy(copytest.DataNamedByteSlice, []byte{9})

	println("OK")
}

// Error:
// cannot copy to readonly tainted slice
