// MAXALLOC: 50000
package main

import "runtime"

var Arr = [3]byte{1, 2, 3}

func xyz() {
	Arr2 := Arr // allocate
	Arr3 := Arr // allocate
	Arr4 := Arr // allocate
	Arr5 := Arr // allocate
	Arr6 := Arr // allocate
	Arr7 := Arr // allocate
	Arr8 := Arr // allocate
	println("memstats in xyz before GC: ", runtime.MemStats())

}
func main() {
	xyz()
	Arr9 := Arr // allocate
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in xyz before GC:  Allocator{maxBytes:50000, bytes:4283}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2606}
