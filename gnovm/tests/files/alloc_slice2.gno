package main

import "runtime"

var S = []string{"a", "b", "c"}

func foo_alloc(i int) int { // note value index
	println(i)
	S1 := S                       // no allocate2
	S2 := []string{"d", "e", "f"} // allocate2
	runtime.GC()
	return i
}

func main() {
	foo_alloc(0)
}
