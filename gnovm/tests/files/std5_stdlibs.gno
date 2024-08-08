package main

import (
	"std"
)

func main() {
	caller1 := std.GetCallerAt(2)
	println(caller1)
}

// Error:
// frame not found
