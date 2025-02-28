// MAXALLOC: 100000000
package main

import "runtime"

func main() {
	var S = [5]int{1, 2, 3, 4, 5}
	S1 := S
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:4200}
