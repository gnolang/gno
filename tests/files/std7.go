package main

import (
	"std"

	"gno.land/p/testutils"
)

func inner() {
	caller1 := std.GetCallerAt(1)
	println(caller1)
	caller2 := std.GetCallerAt(2)
	println(caller2)
	caller3 := std.GetCallerAt(3)
	println(caller3)
}

func main() {
	testutils.WrapCall(inner)
}

// Output:
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
// g1q36x40upm0val7mkzrp5e7a3kxg7cgm548s8v8
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
