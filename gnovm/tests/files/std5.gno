package main

import (
	"std"
)

func main() {
	realm1 := std.CurrentRealm()
	println(realm1)
	realm2 := std.PreviousRealm()
	println(realm2)
}

// Output:
// CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
// UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
