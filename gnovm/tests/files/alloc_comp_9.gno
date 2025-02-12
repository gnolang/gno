// MAXALLOC: 50000
package main

import "runtime"

var s = "hello"

func xyz() {
	for i := 0; i < 100; i++ {
		s += "world!!!"
	}
}
func main() {
	xyz()
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:45900}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2973}
