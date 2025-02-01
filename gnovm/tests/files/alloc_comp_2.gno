// MAXALLOC: 50000

package main

import "runtime"

var a = 1
var S = []string{"a", "b", "c"} // const type

type Foo struct{ name string }

func bar(n int) int {
	clo := func() {
		println("this if funcLit")
	}
	clo2 := clo // no allocation
	ii := new(int)
	// defined within block
	// rhs is struct, array, bigInt, native..
	S1 := S                       // no allocate
	S2 := []string{"d", "e", "f"} // allocate
	var SS []string
	SS = []string{"o", "p", "q"}
	println("memstats in bar before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in bar after GC: ", runtime.MemStats())

	// this should not count
	S3 := []string{"l", "m", "n"}
	S4 := [3]string{"x", "y", "z"}
	S5 := Foo{name: "a"}
	return n + 1
}

func main() {
	type XYZ struct{ name string } // const-type
	bar(1)
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in bar before GC:  Allocator{maxBytes:50000, bytes:4816}
// memstats in bar after GC:  Allocator{maxBytes:50000, bytes:4800}
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:7797}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2664}
