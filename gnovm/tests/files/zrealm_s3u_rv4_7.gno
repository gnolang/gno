// PKGPATH: gno.land/r/zrealm_s3u_rv4_7
// R2a-save rv4: caller's closure run by another realm's helper writes a captured local; caller stores the closure itself
package zrealm_s3u_rv4_7

import "gno.land/r/tests/vm/s3urv"

var fs []func() string

func main(cur realm) {
	x := "v"
	f := func() { x = "w" }
	s3urv.Call(f)
	fs = append(fs, func() string { return x })
	println("x:", fs[0]())
}

// Error:
// gno.land/r/zrealm_s3u_rv4_7: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/zrealm_s3u_rv4_7 and call it with cross)
