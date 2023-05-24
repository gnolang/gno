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

// Error:
// frame not found
