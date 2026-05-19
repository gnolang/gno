// PKGPATH: gno.land/r/alloc7a
// MAXALLOC: 100000000
package alloc7a

import (
	"runtime"

	"gno.land/r/demo/counter"
)

var s = "hello world"

// mainly for inspecting the object load/alloc logic.
func main(cur realm) {
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
	counter.Increment(cross2(cur))
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:8986}
