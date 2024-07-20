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
// return Address<VPBlock(3,22)>(callerAt<VPBlock(3,44)>(n<VPBlock(1,0)>))
//     gonative:std.callerAt
// caller2<VPBlock(1,1)> := std<VPBlock(2,0)>.GetCallerAt((const (2 int)))
//     std/native.gno:44
// main()
//     main/files/std5_stdlibs.gno:10

// Error:
// frame not found
