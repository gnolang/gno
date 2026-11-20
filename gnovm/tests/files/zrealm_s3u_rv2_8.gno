// PKGPATH: gno.land/r/zrealm_s3u_rv2_8
// S3u round 2: an outside relay writes only a never-saved object: declared writer and a self-called stored one.
package zrealm_s3u_rv2_8

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	println("owner:", s3urv.Scratch(cross(cur), s3urv.Stamp2))
}

// Output:
// owner: stamped
