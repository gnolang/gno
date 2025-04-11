package main

import "chain/runtime"

func inner() {
	caller1 := runtime.CallerAt(1)
	println(caller1)
	caller2 := runtime.CallerAt(2)
	println(caller2)
}

func main() {
	inner()
}

// Output:
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
