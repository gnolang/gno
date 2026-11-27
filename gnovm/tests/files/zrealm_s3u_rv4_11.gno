// PKGPATH: gno.land/r/zrealm_s3u_rv4_11
// R2a-save rv4: /p/ counter built and called while the realm is current; once stored it is read-only, so its next call is refused
package zrealm_s3u_rv4_11

import "gno.land/p/demo/tests/entryclosure/v0"

var fs []func() int

func main(cur realm) {
	f := entryclosure.Counter()
	f()
	f()
	fs = append(fs, f)
	println("c:", fs[0]())
}

// Error:
// gno.land/r/zrealm_s3u_rv4_11: an object written by read-only code would be stored (do the write in a crossing function of gno.land/r/zrealm_s3u_rv4_11, called with cross)
