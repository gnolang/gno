// PKGPATH: gno.land/r/zrealm_s3u_rv_4
// S3u review: a fresh victim-owned object returned to the caller is writable
// through the caller's saved closure, though not with its full authority.
package zrealm_s3u_rv_4

import "gno.land/r/tests/vm/s3urv"

var stash = func(e *s3urv.Entry) { e.Owner = "attacker" }

func main(cur realm) {
	e := s3urv.New(cross(cur))
	stash(e)
	s3urv.Keep(cross(cur), e)
	println("owner:", s3urv.Owner(0))
}

// Error:
// gno.land/r/zrealm_s3u_rv_4's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): e<VPBlock(1,0)>.Owner
