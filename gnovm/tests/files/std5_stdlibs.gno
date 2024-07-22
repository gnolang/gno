package main

import (
	"std"
)

func main() {
	caller1 := std.GetCallerAt(1)
	println(caller1)
	caller2 := std.GetCallerAt(2)
	println(caller2)
}

// Stacktrace:

// panic frame not found
// callerAt<VPBlock(3,44)>(n<VPBlock(1,0)>)
//     gonative:std.callerAt
// std<VPBlock(2,0)>.GetCallerAt(2)
//     std/native.gno:44
// main<VPBlock(1,0)>()
//     main/files/std5_stdlibs.gno:10

// Error:
// frame not found
