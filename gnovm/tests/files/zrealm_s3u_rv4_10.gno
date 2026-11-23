// PKGPATH: gno.land/r/zrealm_s3u_rv4_10
// R2a-save rv4: a /p/ counter built through a func value and called by the realm runs as the realm (not view-only); stored fine (cf. rv4_13)
package zrealm_s3u_rv4_10

import "gno.land/p/demo/tests/entryclosure/v0"

var fs []func() int

func main(cur realm) {
	mk := entryclosure.Counter
	f := mk()
	f()
	f()
	fs = append(fs, f)
	println("c:", fs[0]())
}

// Output:
// c: 3
