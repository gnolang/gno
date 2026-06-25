// MAXALLOC: 100000000
package main

import "runtime"

// Exercises the GC recount while a defer is still pending.
//
// The defer is registered inside an inner scope, so that scope's origin block
// is popped (the `}` below) well before f returns. `buf` lives in that popped
// block and is NOT captured by the deferred closure, so it is reachable only
// through the block. A deferred call resolves its own scope from its
// FuncValue (parent + captures), never the origin block, so the popped block
// and its uncaptured locals are provably dead: runtime.GC() must not count
// them. (Counting them — via a Defer.Parent back-reference — would inflate
// bytes by buf's ~64KiB; this golden pins that they are excluded. As with the
// other alloc_* MemStats tests, the exact byte value is calibrated to the full
// TestFiles run, not an isolated -run.)
func f() {
	{
		buf := make([]byte, 1<<16)
		buf[0] = 1
		defer func() { println("defer ran") }()
	}
	runtime.GC()
	println("MemStats: ", runtime.MemStats())
}

func main() {
	f()
}

// Output:
// MemStats:  Allocator{maxBytes:100000000, bytes:7344}
// defer ran
