// MAXALLOC: 50000
package main

import "runtime"

type Foo struct{ name string }

var M = map[string]Foo{
	"a": {name: "a"},
	"b": {name: "b"},
	"c": {name: "c"},
}

func xyz() {
	M["d"] = Foo{"d"}
	M["e"] = Foo{"e"}
	M["f"] = Foo{"f"}
	println("memstats in xyz before GC: ", runtime.MemStats())
}
func main() {
	xyz()
	m := M // no allocation
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in xyz before GC:  Allocator{maxBytes:50000, bytes:5016}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:4464}
