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

// Stacktrace:
// panic: frame not found
// getRealm<VPBlock(3,51)>(1)
//     gonative:std.getRealm
// ref(std).PreviousRealm()
//     std/native.gno:32
// main<VPBlock(1,0)>()
//     main/files/std5.gno:10

// Error:
// frame not found
