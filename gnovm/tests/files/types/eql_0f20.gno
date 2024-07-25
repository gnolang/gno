package main

type f func() bool

// slice would be comparable
var a [2]f

func main() {
	println(a == nil) // rcx.T == nil
}

// Error:
// main/files/types/eql_0f20.gno:9:10: [2]main.f is not comparable
