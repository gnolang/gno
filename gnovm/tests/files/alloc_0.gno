// MAXALLOC: 100000000
package main

import "runtime"

type Foo struct {
	name string
}

var f = Foo{name: "foo"}

func main() {
	f1 := f
	_ = f1 // keep the variable used; the test is about something else
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:8473}
