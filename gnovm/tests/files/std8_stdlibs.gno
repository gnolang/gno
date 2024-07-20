package main

import (
	"std"

	"gno.land/p/demo/testutils"
)

func inner() {
	caller1 := std.GetCallerAt(1)
	println(caller1)
	caller2 := std.GetCallerAt(2)
	println(caller2)
	caller3 := std.GetCallerAt(3)
	println(caller3)
	caller4 := std.GetCallerAt(4)
	println(caller4)
}

func main() {
	testutils.WrapCall(inner)
}

// Stacktrace:
// panic frame not found
// callerAt()
//     std.callerAt
// return Address<VPBlock(3,22)>(callerAt<VPBlock(3,44)>(n<VPBlock(1,0)>))
//     std/native.gno:44
// caller4<VPBlock(1,3)> := std<VPBlock(2,0)>.GetCallerAt((const (4 int)))
//     main/files/std8_stdlibs.gno:16
// fn<VPBlock(1,0)>()
//     gno.land/p/demo/testutils/misc.gno:5
// testutils<VPBlock(2,1)>.WrapCall(inner<VPBlock(3,0)>)
//     main/files/std8_stdlibs.gno:21
// main()
//     main/files/std8_stdlibs.gno:20

// Error:
// frame not found
