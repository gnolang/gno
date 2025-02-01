// MAXALLOC: 50000
package main

import "runtime"

type Foo struct {
	name string
}

var f = Foo{name: "foo"}    // allocate
var pf = &Foo{name: "foo2"} // alloc type pointer, heapItem, struct
var pff = &f

func main() {
	f1 := f  // value copy ,allocate
	f2 := pf // no allocate
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:3128}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:3536}
