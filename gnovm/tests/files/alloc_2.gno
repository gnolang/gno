// MAXALLOC: 100000000
package main

import "runtime"

// GC does not reclaim enough memory
func main() {
	data := make([]byte, 100*1024*1024) // > 100000000
	println("MemStats: ", runtime.MemStats())
}

// Error:
// allocation limit exceeded
