// PKGPATH: gno.land/r/zrealm_s3u_rv3_5
// S3d rv3: the realm's own self-relay of its stored *string writer (S3) on a fresh pointer it stored (cf. rv2_1)
package zrealm_s3u_rv3_5

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	s3urv.SelfPtr(cross(cur))
	println("owner:", s3urv.PtrOwner(0))
}

// Error:
// gno.land/r/tests/vm/s3urv: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/s3urv)
