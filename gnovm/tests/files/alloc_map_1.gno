// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

import "runtime"

var m = map[string]int{"a": 1}

type S struct {
	name string
	mm   map[string]int
}

func bar() {
	var s = S{name: "foo"}
	s.mm = m
	println("MemStats before GC: ", runtime.MemStats())
}

func main() {
	bar()
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())
}

// Output:
// MemStats before GC:  Allocator{maxBytes:9223372036854775807, bytes:2928}
// MemStats after GC:  Allocator{maxBytes:9223372036854775807, bytes:2480}
