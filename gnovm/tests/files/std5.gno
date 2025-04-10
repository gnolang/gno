package main

import "chain/runtime"

func main() {
	caller1 := runtime.CallerAt(1)
	println(caller1)
	caller2 := runtime.CallerAt(2)
	println(caller2)
}

// Stacktrace:
// panic: frame not found
// callerAt<VPBlock(3,50)>(n<VPBlock(1,0)>)
//     gonative:std.callerAt
// std<VPBlock(2,0)>.CallerAt(2)
//     std/native.gno:37
// main<VPBlock(1,0)>()
//     main/files/std5.gno:10

// Error:
// frame not found
