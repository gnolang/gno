package main

type f func() bool

var a [2]f
var b [2]f

// TODO: review where it happens
func main() {
	println(a == nil)
}

// Error:
// main/files/types/0f20_filetest.gno:10: [2]main.f cannot be compared
