// MAXALLOC: 110000000
package main

import "runtime"

// GC does not reclaim enough memory
func main() {
	data := make([]byte, 100*1024*1024) // >100 MB
	data = nil
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())
	data2 := make([]byte, 100*1024*1024) // >100 MB
	println("MemStats after Allocation: ", runtime.MemStats())
}

// Output:
// MemStats after GC:  Allocator{maxBytes:110000000, bytes:2144}
// MemStats after Allocation:  Allocator{maxBytes:110000000, bytes:104862457}
