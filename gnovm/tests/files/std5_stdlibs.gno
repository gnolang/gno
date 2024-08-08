package main

import (
	"std"
)

func main() {
	caller1 := std.GetCallerAt(1)
	println(caller1)
	caller2 := std.GetCallerAt(2)
	println(caller2)
	caller3 := std.GetCallerAt(3)
	println(caller3)
}

// Error:
// frame not found
