// MAXALLOC: 100000000
package main

import "runtime"

var s = "hello world"

func main() {
	s += "!" // underlying array re-allocated
	s1 := s  //  another string value, underlying array same
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:4496}
