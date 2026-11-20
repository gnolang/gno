// PKGPATH: gno.land/r/zrealm_s3u_rv2_6
// S3u round 2: an outside relay writes only never-saved objects: fresh entry and holder.
package zrealm_s3u_rv2_6

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	println("owner:", s3urv.Scratch(cross(cur), s3urv.W))
	println("owner:", s3urv.ScratchHolder(cross(cur), s3urv.HolderW))
}

// Output:
// owner: stamped
// owner: stamped
