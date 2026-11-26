// PKGPATH: gno.land/r/zrealm_s3u_rv2_1
// S3u round 2: the realm's own stored writer (R2a) on a fresh entry in a fresh array attached to stored state
package zrealm_s3u_rv2_1

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	s3urv.SelfRelay(cross(cur))
	println("owner:", s3urv.Owner(0))
}

// Error:
// gno.land/r/tests/vm/s3urv: an object written by read-only code would be stored (do the write in a crossing function of gno.land/r/tests/vm/s3urv, called with cross)
