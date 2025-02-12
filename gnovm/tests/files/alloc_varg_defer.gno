// MAXALLOC: 50000
package main

import "runtime"

func xyz2() {
	defer func(args ...int) { // defer varg
		println(args)
		//runtime.GC() // XXX,GC here does not reclaim memory, but to observe vargs
	}(1, 2, 3)
	runtime.GC() // XXX,GC here does not reclaim memory, but to observe vargs
	println("xyz2, MemStats after GC: ", runtime.MemStats())
}

func main() {
	xyz2()
	println("main, MemStats before GC: ", runtime.MemStats())
	runtime.GC()
	println("main, MemStats after GC: ", runtime.MemStats())
}

// Output:
// xyz2, MemStats after GC:  Allocator{maxBytes:50000, bytes:2600}
// slice[(1 int),(2 int),(3 int)]
// main, MemStats before GC:  Allocator{maxBytes:50000, bytes:5885}
// main, MemStats after GC:  Allocator{maxBytes:50000, bytes:2104}
