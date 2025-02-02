// MAXALLOC: 100000000
package main

import "runtime"

func main() {
	data := map[string]int{"a": 1, "b": 2} // assume this is large enough to trigger GC...
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:2024}
