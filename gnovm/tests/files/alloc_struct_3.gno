// MAXALLOC: 50000
package main

import "runtime"

type Bar struct {
	text string
}

type Foo struct {
	name string
	b    Bar
}

var a = 1
var aa = &a // only alloc new type: pointer type

var f3 = &Foo{name: "ref"}
var f33 = f3

func main() {
	println("MemStats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("MemStats in main after GC: ", runtime.MemStats())
}

// Output:
// MemStats in main before GC:  Allocator{maxBytes:50000, bytes:2640}
// MemStats in main after GC:  Allocator{maxBytes:50000, bytes:3352}
