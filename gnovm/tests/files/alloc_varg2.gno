// MAXALLOC: 50000
package main

import "runtime"

type Foo struct{ n int }

func xyz(args ...Foo) { // varg
	runtime.GC() // XXX, GC here does not reclaim memory, but to observe vargs
	println("xyz, MemStats after GC: ", runtime.MemStats())
}

func main() {
	xyz(Foo{n: 1}, Foo{n: 2}, Foo{n: 3})
	println("main, MemStats before GC: ", runtime.MemStats())
	runtime.GC()
	println("main, MemStats after GC: ", runtime.MemStats())
}

// Output:
// xyz, MemStats after GC:  Allocator{maxBytes:50000, bytes:3752}
// main, MemStats before GC:  Allocator{maxBytes:50000, bytes:5245}
// main, MemStats after GC:  Allocator{maxBytes:50000, bytes:2144}
