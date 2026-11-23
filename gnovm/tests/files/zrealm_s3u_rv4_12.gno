// PKGPATH: gno.land/r/zrealm_s3u_rv4_12
// R2a-save rv4: control for 13: /p/ counter built while the realm is current (heap item stamped), incremented inside another realm's read-only stretch; the realm stores it
package zrealm_s3u_rv4_12

import (
	"gno.land/p/demo/tests/entryclosure/v0"
	"gno.land/r/tests/vm/s3urv"
)

var fs []func() int

func main(cur realm) {
	f := entryclosure.Counter()
	s3urv.Call(func() { f() })
	fs = append(fs, f)
	println("c:", fs[0]())
}

// Error:
// gno.land/r/zrealm_s3u_rv4_12: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/zrealm_s3u_rv4_12)
