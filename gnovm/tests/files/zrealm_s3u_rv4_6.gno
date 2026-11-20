// PKGPATH: gno.land/r/zrealm_s3u_rv4_6
// S3d rv4: caller's closure (made outside the stretch) run by another realm's non-crossing helper writes a captured local; caller stores &x
package zrealm_s3u_rv4_6

import "gno.land/r/tests/vm/s3urv"

var ptrs []*string

func main(cur realm) {
	x := "v"
	s3urv.Call(func() { x = "w" })
	ptrs = append(ptrs, &x)
	println("x:", *ptrs[0])
}

// Error:
// gno.land/r/zrealm_s3u_rv4_6: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/zrealm_s3u_rv4_6)
