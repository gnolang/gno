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
}

func main() {
	testutils.WrapCall(inner)
}

// Output:
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
// g1ssrgtfce6pzs8tp7s6y8473yrffqs9xlgntk0h
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
