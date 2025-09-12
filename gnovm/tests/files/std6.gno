package main

import "chain/runtime"

func inner() {
	realm1 := runtime.CurrentRealm()
	println(realm1)
	realm2 := runtime.PreviousRealm()
	println(realm2)
}

func main() {
	inner()
}

// Output:
// CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
// UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
