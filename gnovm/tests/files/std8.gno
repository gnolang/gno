package main

import (
	"chain/runtime"

	"gno.land/p/demo/testutils"
)

func inner() {
	caller1 := runtime.CallerAt(1)
	println(caller1)
	caller2 := runtime.CallerAt(2)
	println(caller2)
	caller3 := runtime.CallerAt(3)
	println(caller3)
	caller4 := runtime.CallerAt(4)
	println(caller4)
}

func main() {
	testutils.WrapCall(inner)
}

// Stacktrace:
// panic: frame not found
// callerAt<VPBlock(3,11)>(n<VPBlock(1,0)>)
//     gonative:chain/runtime.callerAt
// runtime<VPBlock(2,0)>.CallerAt(4)
//     chain/runtime/native.gno:32
// fn<VPBlock(1,0)>()
//     main/files/std8.gno:16
// testutils<VPBlock(2,1)>.WrapCall(inner<VPBlock(3,0)>)
//     gno.land/p/demo/testutils/misc.gno:5
// main<VPBlock(1,1)>()
//     main/files/std8.gno:21

// Error:
// frame not found
