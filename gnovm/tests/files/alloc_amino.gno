// MAXALLOC: 50000

// PKGPATH: gno.land/r/test
package test

import "runtime"

type Foo struct{ name string }

var f = Foo{name: "foo"}
var m = map[string]Foo{"a": f}

// alloc amino for package block, global objects
func main() {
	println(f, m)
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// (struct{("foo" string)} gno.land/r/test.Foo) map{("a" string):(ref(a8ada09dee16d791fd406d629fe29bb0ed084a30:3) gno.land/r/test.Foo)}
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:45484}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:19750}
