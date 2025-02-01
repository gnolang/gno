// MAXALLOC: 50000
package main

import "runtime"

type Foo struct{ name string }

// S declared in this block then escaped
func xyz() []string {
	var S = []string{"a", "b", "c"}
	for i := 0; i < 100; i++ {
		S = append(S, "d") // A large number of temporary arrays have been deallocated.
	}
	return S
}

func bar(n int) int {
	S1 := xyz()
	// defined within block
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
// MemStats before GC:  Allocator{maxBytes:50000, bytes:25112}
// MemStats after GC:  Allocator{maxBytes:50000, bytes:3360}
