// PKGPATH: gno.land/r/zrealm_s3u_rv2_5
// S3u round 2: the realm's own stored writer (S3) on a holder's fresh nested entry, the holder attached to stored state
package zrealm_s3u_rv2_5

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	s3urv.SelfHolder(cross(cur))
	println("owner:", s3urv.HolderOwner(0))
}

// Error:
// gno.land/r/tests/vm/s3urv: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/s3urv)
