// MAXALLOC: 50000
// PKGPATH: gno.land/r/demo/a/very/long/package/path/so/the/backing/bytes/stand/out/alloc14

// Companion to alloc_14: the same program, but the held strings are fresh
// copies minted by concatenation, so they carry their own backing IDs and
// the GC recount charges their bytes: 40 + 79 = 119 bytes more than
// alloc_14, which holds the handle's field values directly.
package alloc14

import (
	"runtime"
)

func main(cur realm) {
	a := string(cur.Address()) + ""
	p := cur.PkgPath() + ""
	runtime.GC()
	println("===after GC, MemStats: ", runtime.MemStats())
	println(len(a), len(p))
}

// Output:
// ===after GC, MemStats:  Allocator{maxBytes:50000, bytes:9502}
// 40 79
