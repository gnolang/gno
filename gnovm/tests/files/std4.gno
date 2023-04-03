package main

import (
	"std"
)

func main() {
	caller1 := std.GetCallerAt(1)
	println(caller1)
}

// Output:
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
