// PKGPATH: gno.land/r/zrealm_s3u_rv2_3
// S3u round 2: the realm's own stored writer (R2a) on a fresh entry put in a stored map
package zrealm_s3u_rv2_3

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	s3urv.SelfMap(cross(cur))
	println("owner:", s3urv.KeyOwner("k"))
}

// Error:
// gno.land/r/tests/vm/s3urv: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/s3urv)
