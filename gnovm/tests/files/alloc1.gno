// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

import "std"

var c = 1
var b = "hello" // rhs is const
var m = map[string]int{"a": 1}

type Foo struct{ name string }

func bar() {
	f := &Foo{"foo"}
	f2 := f
	println(f)
	bb := &Foo{"bar"}
	println("MemStats 2: ", std.MemStats())
}

func main() {
	a := 1
	println("MemStats 1: ", std.MemStats())
	bar()
	println("MemStats before GC: ", std.MemStats())
	std.GC()
	println("MemStats after GC: ", std.MemStats())
}

// Output:
// MemStats 1:  Allocator{maxBytes:10000000000,bytes,bytes:1904}
// &(struct{("foo" string)} main.Foo)
// MemStats 2:  Allocator{maxBytes:10000000000,bytes,bytes:5984}
// MemStats before GC:  Allocator{maxBytes:10000000000,bytes,bytes:7488}
// MemStats after GC:  Allocator{maxBytes:10000000000,bytes,bytes:3197}
