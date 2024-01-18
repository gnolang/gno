package main

type f func() bool

// slice would be comparable
var a [2]f

func main() {
	println(a == nil) // rcx.T == nil
}

// Error:
// main/files/types/0f20_filetest.gno:9: <nil> and [2]main.f cannot be compared
