// PKGPATH: gno.land/r/zrealm_s3u_rv3_5
// R2a-save rv3: the realm's own self-relay of its stored *string writer (R2a) on a fresh pointer it stored (cf. rv2_1)
package zrealm_s3u_rv3_5

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	s3urv.SelfPtr(cross(cur))
	println("owner:", s3urv.PtrOwner(0))
}

// Error:
// gno.land/r/tests/vm/s3urv: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/tests/vm/s3urv and call it with cross)
