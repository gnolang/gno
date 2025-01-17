package main

import "runtime"

var f = func() bool { return true } // allocate2

func main() {
	f1 := f // ??? why not value copy?
	runtime.GC()
}
