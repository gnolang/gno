// MAXALLOC: 100000000
package main

import "runtime"

type Foo struct {
	name string
}

var f = Foo{name: "foo"}

func main() {
	f1 := f
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:4862}
