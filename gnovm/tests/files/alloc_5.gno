// MAXALLOC: 100000000
package main

import "runtime"

func gen() {
	data := make([]byte, 50*1024*1024)
}

// this is mainly for gas usage check
// see gc.txtar
func main() {
	for i := 0; i < 10; i++ {
		gen()
		gen()
	}
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main after GC:  Allocator{maxBytes:100000000, bytes:6048}

// TypeCheckError:
// main/files/alloc_5.gno:7:2: declared and not used: data
