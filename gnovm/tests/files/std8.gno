package main

import (
	"std"

	"gno.land/p/demo/testutils"
)

func inner() {
	caller1 := std.CallerAt(1)
	println(caller1)
	caller2 := std.CallerAt(2)
	println(caller2)
	caller3 := std.CallerAt(3)
	println(caller3)
	caller4 := std.CallerAt(4)
	println(caller4)
}

func main() {
	testutils.WrapCall(inner)
}

// Error:
// frame not found
