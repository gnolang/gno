package main

import "std"

func inner() {
	realm1 := std.CurrentRealm()
	println(realm1)
	realm2 := std.PreviousRealm()
	println(realm2)
}

func main() {
	inner()
}

// Stacktrace:
// panic: frame not found
// getRealm<VPBlock(3,49)>(1)
//     gonative:std/std.gno
// ref(std).PreviousRealm()
//     std/native.gno:32
// inner<VPBlock(3,0)>()
//     main/files/std6.gno:8
// main<VPBlock(1,1)>()
//     main/files/std6.gno:13

// Error:
// frame not found
