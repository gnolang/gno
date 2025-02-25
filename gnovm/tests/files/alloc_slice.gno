// MAXALLOC: 100000000
package main

import "runtime"

//var S = []int{1, 2, 3, 4, 5}

// GC does not reclaim enough memory
func main() {
	var S = []int{1, 2, 3, 4, 5}
	S1 := S
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:4432}
