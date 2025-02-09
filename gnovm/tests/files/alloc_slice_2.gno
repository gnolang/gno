// MAXALLOC: 50000
package main

import "runtime"

var S = []string{"a", "b", "c"} // this remains in memory

func xyz() {
	for i := 0; i < 100; i++ {
		S = append(S, "d") // A large number of temporary arrays have been deallocated.
	}
}

func bar(n int) {
	xyz()
	println("MemStats in bar before GC: ", runtime.MemStats())
	S1 := S[1:] // no allocation
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
// MemStats in bar before GC:  Allocator{maxBytes:50000, bytes:43928}
// MemStats in bar after GC:  Allocator{maxBytes:50000, bytes:11560}
// MemStats in main before GC:  Allocator{maxBytes:50000, bytes:13054}
// MemStats in main after GC:  Allocator{maxBytes:50000, bytes:6584}
