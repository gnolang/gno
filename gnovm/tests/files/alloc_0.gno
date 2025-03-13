// MAXALLOC: 100000000
package main

import "runtime"

type Foo struct {
	name string
}

// how is type reused?
var f = Foo{name: "foo"}

func main() {
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:4512}
