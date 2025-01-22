// MAXALLOC: 100

package main

import "runtime"

var a = 1
var S = []string{"a", "b", "c"} // const type

type Foo struct{ name string }

func bar(n int) int {
	clo := func() {
		println("this if funcLit")
	}
	ii := new(int)
	// defined within block
	// rhs is struct, array, bigInt, native..
	S1 := S                       // no allocate2
	S2 := []string{"d", "e", "f"} // allocate2
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
// memstats in bar before GC:  Allocator{maxBytes:9223372036854775807, bytes:4776}
// memstats in bar after GC:  Allocator{maxBytes:9223372036854775807, bytes:4760}
// memstats in main before GC:  Allocator{maxBytes:9223372036854775807, bytes:7771}
// memstats in main after GC:  Allocator{maxBytes:9223372036854775807, bytes:2664}
