package main

type f func() bool

var a f = func() bool { return true }
var b f = func() bool { return false }

func main() {
	println(a == b)
}

// Error:
// main/files/types/eql_0f17.gno:9:10: main.f can only be compared to nil

// TypeCheckError:
// main/files/types/eql_0f17.gno:9:10: invalid operation: a == b (func can only be compared to nil)
