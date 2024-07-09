package main

type f func()

var f1 f
var f2 f

func main() {
	println(f1 == f2)
}

// Error:
// main/files/types/eql_0f30b.gno:9:10: main.f can only be compared to nil
