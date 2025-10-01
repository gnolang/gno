// MAXALLOC: 50000
package main

import "runtime"

var s = "hello"

func xyz() {
	for i := 0; i < 110; i++ {
		s += "world!!!"
	}
}
func main() {
	xyz() // first gc happens in xyz(), and m.Package is main
	println("memstats in main after first GC: ", runtime.MemStats())

	// first GC happens in xyz(), then more allocations
	// happens.
	// after second gc, more garbage collected.
	runtime.GC()
	println("memstats in main after second GC: ", runtime.MemStats())
}

// Output:
// memstats in main after first GC:  Allocator{maxBytes:50000, bytes:16511}
// memstats in main after second GC:  Allocator{maxBytes:50000, bytes:6901}
