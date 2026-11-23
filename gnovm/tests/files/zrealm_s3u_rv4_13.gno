// PKGPATH: gno.land/r/zrealm_s3u_rv4_13
// R2a-save rv4: /p/ counter built through a func value (view-only: zero-stamp heap item), incremented inside another realm's read-only stretch; the realm stores it
package zrealm_s3u_rv4_13

import (
	"gno.land/p/demo/tests/entryclosure/v0"
	"gno.land/r/tests/vm/s3urv"
)

var fs []func() int

func main(cur realm) {
	mk := entryclosure.Counter
	f := mk()
	s3urv.Call(func() { f() })
	fs = append(fs, f)
	println("c:", fs[0]())
}

// Error:
// gno.land/r/zrealm_s3u_rv4_13: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/zrealm_s3u_rv4_13)
