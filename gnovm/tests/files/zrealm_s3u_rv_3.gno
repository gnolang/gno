// PKGPATH: gno.land/r/zrealm_s3u_rv_3
// S3u review: no pointer handed over; the hook reads the victim's stored
// (but not yet finalized) entry and writes it through a saved closure.
package zrealm_s3u_rv_3

import "gno.land/r/tests/vm/s3urv"

var stash = func() { s3urv.Last().Owner = "attacker" }

func OnAdd(cur realm) { stash() }

func main(cur realm) {
	s3urv.AddNotify(cross(cur), OnAdd)
	println("owner:", s3urv.Owner(0))
}

// Error:
// gno.land/r/zrealm_s3u_rv_3's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/s3urv) package{})).Last().Owner
