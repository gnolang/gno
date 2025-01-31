// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

import "runtime"

var m = map[string]int{"a": 1}

type S struct {
	name string
	mm   map[string]int
}

func main() {
	var s = S{name: "foo"}
	s.mm = m // no allocation
	println("MemStats before GC: ", runtime.MemStats())
	runtime.GC()
	// NOTE: After garbage collection (GC), the memory usage is 2744 bytes,
	// compared to 2432 bytes before GC. This is due to two factors:
	// 1. The GC does not release memory because the block was not popped off the heap.
	// 2. The GC process itself introduces some additional memory overhead.
	println("MemStats after GC: ", runtime.MemStats())
}

// Output:
// MemStats before GC:  Allocator{maxBytes:9223372036854775807, bytes:2432}
// MemStats after GC:  Allocator{maxBytes:9223372036854775807, bytes:2744}
