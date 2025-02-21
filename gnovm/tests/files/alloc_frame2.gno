// MAXALLOC: 50000
package main

import "runtime"

type foo struct {
	n int
}

type ft = func()

var fns []ft

func xyz() {
	for i := 0; i < 5; i++ {
		z := i
		var fv = func() {
			a := foo{n: z}
			println(a.n)
			runtime.GC()
		}
		fns = append(fns, fv)
	}
}

func main() {
	xyz()
	for _, fn := range fns {
		fn()
	}
}

// Output:
// 0
// 1
// 2
// 3
// 4
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:20136}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:5064}
