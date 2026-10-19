// MAXALLOC: 110000000
package main

import "runtime"

func main() {
	data := make([]byte, 100*1024*1024) // < 110000000
	println(len(data))                  // a use; `_ = data` would pin the slice in the block's blank slot
	data = nil
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())
}

// Output:
// 104857600
// MemStats after GC:  Allocator{maxBytes:110000000, bytes:7232}
