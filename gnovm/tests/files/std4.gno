package main

import (
	"chain/runtime/unsafe"
)

func main() {
	realm1 := unsafe.CurrentRealm()
	println(realm1)
}

// Output:
// CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
