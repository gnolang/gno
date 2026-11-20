// PKGPATH: gno.land/r/zrealm_s3u_rv4_11
// S3d rv4: control: /p/ counter built and called while the realm is current; the realm stores it
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

// Output:
// c: 3
