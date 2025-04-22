package main

type f func() bool

var a [2]f
var b [2]f

func main() {
	println(a == b)
}

// Error:
// main/files/types/eql_0f21.gno:9:10: [2]main.f is not comparable

// TypeCheckError:
// main/files/types/eql_0f21.gno:9:10: invalid operation: a == b ([2]f cannot be compared)
