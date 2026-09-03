// MAXALLOC: 50000
// PKGPATH: gno.land/r/demo/a/very/long/package/path/so/the/backing/bytes/stand/out/alloc14

// realm.Address() and realm.PkgPath() return the realm handle's own field
// StringValues, not fresh copies, so the results share the handle's
// backing IDs. In the GC recount each result adds a string header only,
// never its backing bytes again.
//
// Companion alloc_14a holds fresh copies of the same two strings (minted
// by concatenation) and lands exactly len(a)+len(p) = 40+79 = 119 bytes
// higher after GC: that gap is the dedup. Companion alloc_14b lets the
// handle die first, so only tracked results keep their bytes in the
// recount: that is the tracking. An untracked (ID 0) result passes here
// but fails alloc_14b by exactly its length.
package alloc14

import (
	"runtime"
)

func main(cur realm) {
	a := cur.Address()
	p := cur.PkgPath()
	runtime.GC()
	println("===after GC, MemStats: ", runtime.MemStats())
	println(len(a), len(p))
}

// Output:
// ===after GC, MemStats:  Allocator{maxBytes:50000, bytes:9383}
// 40 79
