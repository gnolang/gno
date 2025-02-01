// MAXALLOC: 50000
package main

import "runtime"

func main() {
	var s = "hello" // const
	s += "world"
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:1642}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2138}
