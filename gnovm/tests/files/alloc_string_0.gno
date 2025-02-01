// MAXALLOC: 50000
package main

import "runtime"

var s = "hello"

func main() {
	s += "world" // new string
	s += "!!!"   // new string
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:1639}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2141}
