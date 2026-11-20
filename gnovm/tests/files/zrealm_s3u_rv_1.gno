// PKGPATH: gno.land/r/zrealm_s3u_rv_1
// S3u review: victim crosses into the attacker's hook with a fresh entry it
// already stored; the attacker routes the write through its own saved closure.
package zrealm_s3u_rv_1

import "gno.land/r/tests/vm/s3urv"

var stash = func(e *s3urv.Entry) { e.Owner = "attacker" } // made at init

func OnAdd(cur realm, e *s3urv.Entry) { stash(e) }

func main(cur realm) {
	s3urv.Add(cross(cur), OnAdd)
	println("owner:", s3urv.Owner(0))
}

// Error:
// gno.land/r/zrealm_s3u_rv_1's code is read-only here: a function value writes only in the crossing call that made it: e<VPBlock(1,0)>.Owner
