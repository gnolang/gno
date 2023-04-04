package main

import "std"

func inner() {
	caller1 := std.GetCallerAt(1)
	println(caller1)
	caller2 := std.GetCallerAt(2)
	println(caller2)
}

func main() {
	inner()
}

// Output:
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
