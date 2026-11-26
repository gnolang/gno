// PKGPATH: gno.land/r/zrealm_s3u_rv_2
// S3u review control: the same write with the attacker's full authority.
package zrealm_s3u_rv_2

import "gno.land/r/tests/vm/s3urv"

func OnAdd(cur realm, e *s3urv.Entry) { e.Owner = "attacker" }

func main(cur realm) {
	s3urv.Add(cross(cur), OnAdd)
	println("owner:", s3urv.Owner(0))
}

// Error:
// cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm; a /p/ package's object cannot be written after its init): e<VPBlock(1,1)>.Owner
