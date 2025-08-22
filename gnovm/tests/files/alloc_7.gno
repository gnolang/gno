// MAXALLOC: 100000000
package main

import (
	"runtime"

	"gno.land/r/demo/ext"
)

var s = "hello world" // NOTE, this counts twice for now, should improve.

func main() {
	s += "!" // underlying array re-allocated
	s1 := s  //  another string value, underlying array same
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
	ext.GetVar()
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:7128}

// TypeCheckError:
// main/alloc_7.gno:14:2: declared and not used: s1
