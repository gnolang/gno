// MAXALLOC: 100000000
package main

import "runtime"

var s = "hello world" // NOTE, this counts twice for now, should improve.

func main() {
	s += "!" // underlying array re-allocated
	s1 := s  //  another string value, underlying array same
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:5876}

// TypeCheckError:
// main/alloc_7.gno:10:2: declared and not used: s1
