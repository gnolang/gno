package main

import (
	"std"
)

func main() {
	caller1 := std.CallerAt(1)
	println(caller1)
	caller2 := std.CallerAt(2)
	println(caller2)
}

// Stacktrace:
// panic: frame not found
// callerAt<VPBlock(3,52)>(n<VPBlock(1,0)>)
//     gonative:std.callerAt
// std<VPBlock(2,0)>.CallerAt(2)
//     std/native.gno:41
// main<VPBlock(1,0)>()
//     main/files/std5.gno:10

// Error:
// panic: frame not found
