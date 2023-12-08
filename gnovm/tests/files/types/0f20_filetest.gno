package main

type f func() bool

var a [2]f
var b [2]f

func main() {
	println(a == nil) // rcx.T == nil
}

// Error:
// main/files/types/0f20_filetest.gno:9: [2]main.f cannot be compared
