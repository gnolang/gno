package main

import "std"

func main() {
	caller := std.GetOrigCaller()
	println(caller)
}

// Output:
// g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
