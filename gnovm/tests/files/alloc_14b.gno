// MAXALLOC: 50000
// PKGPATH: gno.land/r/demo/a/very/long/package/path/so/the/backing/bytes/stand/out/alloc14

// Companion to alloc_14 (see its header): the handle dies before the GC
// (grab's cur is gone once it returns) and only the results survive, so
// their bytes stay in the recount only if they are tracked.
package alloc14

import (
	"runtime"
)

var (
	a address
	p string
)

func grab(cur realm) {
	a = cur.Address()
	p = cur.PkgPath()
}

func main(cur realm) {
	grab(cross(cur))
	runtime.GC()
	println("===after GC, MemStats: ", runtime.MemStats())
	println(len(a), len(p))
}

// Output:
// ===after GC, MemStats:  Allocator{maxBytes:50000, bytes:11478}
// 40 79
