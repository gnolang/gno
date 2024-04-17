package main

type f func() bool

var a [2]f
var b [2]f

func main() {
	println(a == b)
}

// Error:
// main/files/types/0f21_filetest.gno:9: [2]main.f is not comparable
