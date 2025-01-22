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
	runtime.GC()
	println(runtime.MemStats())

	// this should not count
	S3 := []string{"l", "m", "n"}
	S4 := [3]string{"x", "y", "z"}
	S5 := Foo{name: "a"}
	return n + 1
}

func main() {
	type XYZ struct{ name string } // const-type
	bar(1)
}

// Output:
// Allocator{maxBytes:9223372036854775807, bytes:4760}
