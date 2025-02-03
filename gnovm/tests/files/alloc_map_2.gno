// MAXALLOC: 50000
package main

import "runtime"

type Foo struct{ name string }

func xyz() {
	var M = map[string][]Foo{}
	M["a"] = []Foo{Foo{"a"}}
	M["b"] = []Foo{Foo{"b"}}
	M["c"] = []Foo{Foo{"c"}}

	println("memstats in xyz before GC: ", runtime.MemStats())
}
func main() {
	xyz()
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in xyz before GC:  Allocator{maxBytes:50000, bytes:4272}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2144}
