// MAXALLOC: 50000
package main

import "runtime"

func main() {
	s := "hello" + "world" // const, no alloc runtime
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:1608}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2104}
