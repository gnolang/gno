package main

import (
	"std"
)

func main() {
	realm1 := std.CurrentRealm()
	println(realm1)
}

// Output:
// (struct{("g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4" std.Address),("main" string)} std.Realm)
