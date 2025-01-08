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
	println(f)
	bb := &Foo{"bar"}
}

func main() {
	a := 1
	bar()
	println("MemStats before GC: ", std.MemStats())
	std.GC()
	println("MemStats after GC: ", std.MemStats())
}

// Output:
// &(struct{("foo" string)} main.Foo)
// MemStats before GC:  Allocator{maxBytes:10000000000,bytes,bytes:4440}
// MemStats after GC:  Allocator{maxBytes:10000000000,bytes,bytes:3197}
