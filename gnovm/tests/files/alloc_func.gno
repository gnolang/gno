// MAXALLOC: 50000
package main

import "runtime"

var f = func() bool { return true } // alloc

func main() {
	f1 := f // ??? why not value copy?
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:1776}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2312}
