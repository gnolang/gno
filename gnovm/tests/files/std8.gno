package main

import (
	"std"

	"gno.land/p/demo/testutils"
)

func inner() {
	realm1 := std.CurrentRealm()
	println(realm1)
	realm2 := std.PreviousRealm()
	println(realm2)
}

func main() {
	testutils.WrapCall(inner)
}

// Output:
// (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)

// Stacktrace:
// panic: frame not found
// getRealm<VPBlock(3,49)>(1)
//     gonative:std/std.gno
// ref(std).PreviousRealm()
//     std/native.gno:32
// fn<VPBlock(1,0)>()
//     main/files/std8.gno:12
// ref(gno.land/p/demo/testutils).WrapCall(inner<VPBlock(3,0)>)
//     gno.land/p/demo/testutils/misc.gno:5
// main<VPBlock(1,1)>()
//     main/files/std8.gno:17

// Error:
// frame not found
