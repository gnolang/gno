// MAXALLOC: 100000000
package main

import (
	"runtime"

	"gno.land/r/demo/counter"
)

var s = "hello world"

// mainly for inspecting the object load/alloc logic.
func main() {
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
	counter.Increment(cross)
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:8077}
