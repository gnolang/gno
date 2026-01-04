// MAXALLOC: 110000000
package main

import "runtime"

func main() {
	data := make([]byte, 100*1024*1024) // < 110000000
	data = nil
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())
}

// Output:
// MemStats after GC:  Allocator{maxBytes:110000000, bytes:5744}

// TypeCheckError:
// main/alloc_3.gno:7:2: declared and not used: data
