package main

import (
	"std"
)

func main() {
	caller1 := std.GetCallerAt(1)
	println(caller1)
	caller2 := std.GetCallerAt(2)
	println(caller2)
}

// Error:
// panic: frame not found
