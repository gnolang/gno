// MAXALLOC: 25000
// max total allocation of 100 mb.

// PKGPATH: gno.land/r/test
package test

import "runtime"

type Foo struct{ name string }

var f = Foo{name: "foo"}
var m = map[string]Foo{"a": f}

func main() {
	println(f, m) // here should alloc amino for f
	runtime.GC()
}
