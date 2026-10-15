// MAXALLOC: 100000000
package main

import "runtime"

// GC does not reclaim enough memory
func main() {
	data := make([]byte, 100*1024*1024) // > 100000000
	_ = data                            // keep the variable used; the test is about something else
	println("MemStats: ", runtime.MemStats())
}

// Error:
// allocation limit exceeded
