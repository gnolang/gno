package main

import "runtime"

var S = []string{"a", "b", "c"}

func main() {
	S1 := S                       // no allocate2
	S2 := []string{"d", "e", "f"} // allocate2
	runtime.GC()
}
