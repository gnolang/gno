package main

import "fmt"

// TODO: the log is runtime log, implies it escape preprocess check
func foo(a uint64, b float32) {
	fmt.Printf("%T \n", a)
	println(a)
}

func main() {
	x := 11
	foo(1<<x, 1>>x)
}

// Error:
// main/files/types2/10a17b.gno:13: operator >> not defined on: float32
