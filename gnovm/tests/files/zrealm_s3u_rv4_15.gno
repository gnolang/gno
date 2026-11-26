// PKGPATH: gno.land/r/zrealm_s3u_rv4_15
// R2a-save rv4: as rv4_14, but the cursor is advanced inside a second read-only stretch (another realm's helper), then the iterator is stored
package zrealm_s3u_rv4_15

import (
	"gno.land/p/moul/fifo/v0"
	"gno.land/r/tests/vm/s3urv"
)

var l = fifo.New(10)

var its []func() any

func main(cur realm) {
	l.Append("a")
	l.Append("b")
	l.Append("c")
	mv := l.Iterator
	var it func() any
	s3urv.Call(func() { it = mv() })
	s3urv.Call(func() { println("1:", it()) })
	its = append(its, it)
	println("stored")
}

// Error:
// gno.land/r/zrealm_s3u_rv4_15: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/zrealm_s3u_rv4_15 and call it with cross)
