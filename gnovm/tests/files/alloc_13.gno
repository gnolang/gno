// MAXALLOC: 10000
// PKGPATH: gno.land/r/test

// Tests string byte recounting across two GC cycles.
//
// Two GCs happen in this test:
//  1. Auto-GC — MAXALLOC=10000 is tight, so the allocator triggers GC
//     automatically somewhere in the preamble.
//  2. runtime.GC() in main() — explicit recount; verifies that string
//     underlying bytes are correctly recounted in a second cycle.
//
// s1 := s shares the same Go backing array. Within a single GC cycle the
// backing bytes are counted only once (the second visit adds just the
// StringValue header). Across cycles each cycle recounts the full string.
//
// Memory formula (approximate):
//
//	before GC = sum of all live objects after auto-GC
//	after GC  = before GC + new allocations (e.g. s1 header, print strings)
//	           - dead objects freed by GC
package test

import (
	"runtime"
)

var s string

func Assign() {
	s = `aaaaaaaaaaaaaaaaaaaa`
}

func init() {
	Assign()
}

func main() {
	s1 := s // shares backing with s; GC recounts each backing once per cycle
	println("===before GC, MemStats: ", runtime.MemStats())
	runtime.GC()
	println("===after GC, MemStats: ", runtime.MemStats())
	_ = s1
}

// Output:
// ===before GC, MemStats:  Allocator{maxBytes:10000, bytes:6956}
// ===after GC, MemStats:  Allocator{maxBytes:10000, bytes:7692}
