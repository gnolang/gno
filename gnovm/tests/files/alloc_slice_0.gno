// MAXALLOC: 50000
package main

import "runtime"

var S = []string{"a", "b", "c"} // this remains in memory

type Foo struct{ name string }

func xyz() {
	for i := 0; i < 100; i++ {
		S = append(S, "d") // A large number of temporary arrays have been deallocated.
	}
}

func bar(n int) int {
	xyz() // after this, S still in memory
	// defined within block
	S1 := S                       // no allocate
	S2 := []string{"d", "e", "f"} // allocate
	println("MemStats before GC: ", runtime.MemStats())
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())

	// this should not count
	S3 := []string{"l", "m", "n"}
	S4 := [3]string{"x", "y", "z"}
	S5 := Foo{name: "a"}
	return n + 1
}

func main() {
	bar(1)
}

// Output:
// MemStats before GC:  Allocator{maxBytes:50000, bytes:25072}
// MemStats after GC:  Allocator{maxBytes:50000, bytes:7800}
