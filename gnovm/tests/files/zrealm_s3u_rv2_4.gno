// PKGPATH: gno.land/r/zrealm_s3u_rv2_4
// S3u round 2: the realm's own stored writer (R2a) on a fresh entry attached then detached, then kept
package zrealm_s3u_rv2_4

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	s3urv.SelfDetach(cross(cur))
	println("owner:", s3urv.KeyOwner("t"))
}

// Error:
// gno.land/r/tests/vm/s3urv: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/s3urv)
