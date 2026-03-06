package main

import "chain/runtime"

func main() {
	realm1 := runtime.CurrentRealm()
	println(realm1)
}

// Output:
// CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
