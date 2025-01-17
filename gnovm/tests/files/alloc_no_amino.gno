// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

import "runtime"

type Foo struct{ name string }

var f = Foo{name: "foo"}

func main() {
	println(f) // here should alloc amino for f
	runtime.GC()
}
