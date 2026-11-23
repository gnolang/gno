// PKGPATH: gno.land/r/zrealm_s3u_rv3_6
// R2a-save rv3: an outside relay of the victim's stored *string writer onto the victim's fresh stored pointer (cf. rv2_6)
package zrealm_s3u_rv3_6

import "gno.land/r/tests/vm/s3urv"

func main(cur realm) {
	s3urv.AddPtr(cross(cur), s3urv.WS)
	println("owner:", s3urv.PtrOwner(0))
}

// Error:
// gno.land/r/tests/vm/s3urv: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/s3urv)
