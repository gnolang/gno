package main

import "runtime"

var Arr = [3]string{"a", "b", "c"} // allocate2

func main() {
	Arr2 := Arr // allocate2
	runtime.GC()
}
