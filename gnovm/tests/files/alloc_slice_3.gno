// MAXALLOC: 50000
package main

import "runtime"

var arr [103]string

func init() {
	arr[0] = "a"
	arr[1] = "b"
	arr[2] = "c"

	for i := 3; i < len(arr); i++ {
		arr[i] = "d"
	}
}

func bar(n int) {
	println("MemStats in bar before GC: ", runtime.MemStats())
	S1 := arr[1:] // no allocation
	runtime.GC()
	println("MemStats in bar after GC: ", runtime.MemStats())
}

func main() {
	bar(1)
	println("MemStats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("MemStats in main after GC: ", runtime.MemStats())
}

// Output:
// MemStats in bar before GC:  Allocator{maxBytes:50000, bytes:18544}
// MemStats in bar after GC:  Allocator{maxBytes:50000, bytes:7392}
// MemStats in main before GC:  Allocator{maxBytes:50000, bytes:8885}
// MemStats in main after GC:  Allocator{maxBytes:50000, bytes:2184}
