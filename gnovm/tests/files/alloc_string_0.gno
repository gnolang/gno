// MAXALLOC: 25000
// max total allocation of 100 mb.

package main

import "runtime"

var s = "hello"

func main() {
	s += "world" // new string
	s += "!!!"   // new string
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:9223372036854775807, bytes:1639}
// memstats in main after GC:  Allocator{maxBytes:9223372036854775807, bytes:2141}
