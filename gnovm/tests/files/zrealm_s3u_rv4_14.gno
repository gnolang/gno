// PKGPATH: gno.land/r/zrealm_s3u_rv4_14
// R2a-save rv4: a /p/ iterator built by a method value inside another realm's read-only stretch (zero-stamp heap item for its cursor); later view-only calls advance the cursor; the realm stores the iterator
package zrealm_s3u_rv4_14

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
	println("1:", it())
	println("2:", it())
	its = append(its, it)
	println("stored")
}

// Error:
// gno.land/r/zrealm_s3u_rv4_14: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/zrealm_s3u_rv4_14 and call it with cross)
