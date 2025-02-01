// MAXALLOC: 100000000
package main

import "runtime"

func main() {
	data := make([]byte, 100*1024*1024) // 100 MB
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:2616}
