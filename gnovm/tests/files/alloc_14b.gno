// MAXALLOC: 50000
// PKGPATH: gno.land/r/demo/a/very/long/package/path/so/the/backing/bytes/stand/out/alloc14

// Companion to alloc_14: the realm handle whose fields Address()/PkgPath()
// return dies before the GC (grab's cur is gone once it returns), and only
// the results survive in globals. Because the results carry the handle's
// mint IDs the recount still charges their backing bytes (40 + 79 = 119).
// An untracked (ID 0) copy would drop out of the recount and the golden
// would fall by the missing bytes.
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
