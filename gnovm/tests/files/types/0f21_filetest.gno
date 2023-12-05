package main

type f func() bool

var a [2]f
var b [2]f

// TODO: review where it happens
func main() {
	println(a == b)
}

// Error:
// main/files/types/0f21_filetest.gno:10: main.f cannot be compared
