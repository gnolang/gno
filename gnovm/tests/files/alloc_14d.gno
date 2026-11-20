// MAXALLOC: 100000
// PKGPATH: gno.land/r/test
// s+"" and ""+s keep s's backing, so the GC recount charges s once.
package test

import (
	"runtime"
	"strings"
)

var a, b, c string

func main() {
	a = strings.Repeat("x", 10000)
	b = a + ""
	c = "" + a
	runtime.GC()
	println(runtime.MemStats())
}

// Output:
// Allocator{maxBytes:100000, bytes:25528}
