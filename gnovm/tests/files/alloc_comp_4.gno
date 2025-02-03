// MAXALLOC: 100000000
package main

import "runtime"

// GC cannot reclaim enough memory
func main() {
	data := make([]byte, 100*1024*1024) // >100 MB
	println("MemStats: ", runtime.MemStats())
}

// Error:
// allocation limit exceeded
