// MAXALLOC: 50000
package main

import "runtime"

func xyz(args ...int) { // varg
	runtime.GC() // XXX, GC here does not reclaim memory, but to observe vargs
	println("xyz, MemStats after GC: ", runtime.MemStats())
}

func main() {
	xyz(1, 2, 3)
	println("main, MemStats before GC: ", runtime.MemStats())
	runtime.GC()
	println("main, MemStats after GC: ", runtime.MemStats())
}

// Output:
// xyz, MemStats after GC:  Allocator{maxBytes:50000, bytes:3040}
// main, MemStats before GC:  Allocator{maxBytes:50000, bytes:4533}
// main, MemStats after GC:  Allocator{maxBytes:50000, bytes:2104}
