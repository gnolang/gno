// PKGPATH: gno.land/r/zrealm_s3u_rv2_2
// S3u round 2: the realm's own stored writer (R2a) on a fresh entry nested in a fresh holder attached to stored state
package zrealm_s3u_rv2_2

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	s3urv.SelfNested(cross(cur))
	println("owner:", s3urv.HolderOwner(0))
}

// Error:
// gno.land/r/tests/vm/s3urv: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/tests/vm/s3urv and call it with cross)
