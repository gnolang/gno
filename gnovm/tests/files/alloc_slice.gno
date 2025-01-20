package main

import "runtime"

var S = []string{"a", "b", "c"}

type Foo struct{ name string }

func bar(n int) int {
	// defined within block
	// rhs is struct, array, bigInt, native..
	S1 := S                       // no allocate2
	S2 := []string{"d", "e", "f"} // allocate2
	runtime.GC()

	// this should not count
	S3 := []string{"l", "m", "n"}
	S4 := [3]string{"x", "y", "z"}
	S5 := Foo{name: "a"}
	return n + 1
}

func main() {
	bar(1)
}
