package main

import (
	"chain/runtime/unsafe"

	"gno.land/p/nt/testutils/v0"
)

func inner() {
	realm1 := unsafe.CurrentRealm()
	println(realm1)
	realm2 := unsafe.PreviousRealm()
	println(realm2)
}

func main() {
	testutils.WrapCall(inner)
}

// Output:
// CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
// UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
