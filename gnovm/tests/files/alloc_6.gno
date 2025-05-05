// MAXALLOC: 100000000
package main

import "runtime"

func main() {
	var a = func() int {
		return 1
	}

	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main after GC:  Allocator{maxBytes:100000000, bytes:6048}

// TypeCheckError:
// main/files/alloc_6.gno:7:6: declared and not used: a
