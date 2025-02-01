// MAXALLOC: 50000
package main

import "runtime"

func main() {
	w := "world"
	s := "hello" + w
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:1682}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2178}
