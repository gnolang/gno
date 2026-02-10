// MAXALLOC: 100000000
package main

import "runtime"

type Bar struct {
	name string
}

type Foo struct {
	name string
	Bar
}

var S = []*Foo{&Foo{name: "foo1", Bar: Bar{name: "bar1"}}, &Foo{name: "foo2", Bar: Bar{name: "bar2"}}}

func main() {
	S1 := S
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:7648}

// TypeCheckError:
// main/alloc_1.gno:18:2: declared and not used: S1
