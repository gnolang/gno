package main

type f func()

var f1 f
var f2 f

func main() {
	println(f1 == f2)
}

// Error:
// main/eql_0f30b.gno:9:10-18: main.f can only be compared to nil

// TypeCheckError:
// main/eql_0f30b.gno:9:10: invalid operation: f1 == f2 (func can only be compared to nil)
