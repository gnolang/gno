// MAXALLOC: 100000000
package main

import "runtime"

// GC does not reclaim enough memory
func main() {
	var a = func() int {
		return 1
	}

	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main after GC:  Allocator{maxBytes:100000000, bytes:4604}
