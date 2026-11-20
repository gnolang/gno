// PKGPATH: gno.land/r/zrealm_s3u_rv2_7
// S3u round 2 positive control: declared writer and a self-called stored one.
package zrealm_s3u_rv2_7

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	println("owner:", s3urv.ScratchSelf(cross(cur)))
}

// Output:
// owner: stamped
