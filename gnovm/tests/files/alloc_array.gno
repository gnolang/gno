// MAXALLOC: 50000
package main

import "runtime"

var Arr = [3]string{"a", "b", "c"} // allocate2

func main() {
	Arr2 := Arr // allocate
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:2592}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2800}
