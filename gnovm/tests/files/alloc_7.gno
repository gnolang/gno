// MAXALLOC: 100000000
package main

import (
	"runtime"
)

var s = "hello world"

func main() {
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:5851}
