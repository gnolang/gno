// MAXALLOC: 110000000
package main

import "runtime"

func main() {
	data := make([]byte, 100*1024*1024) // >100 MB
	data = nil
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())
}

// Output:
// MemStats after GC:  Allocator{maxBytes:110000000, bytes:4112}
