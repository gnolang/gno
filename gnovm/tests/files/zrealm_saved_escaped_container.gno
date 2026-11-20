// PKGPATH: gno.land/r/zrealm_saved_escaped_container
// S3 attack: R.Do builds a local table and a caller's crossing callback
// re-enters R.RunAll with it; the fresh entry must not write.

package zrealm_saved_escaped_container

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.Do(cross(cur), func(cur realm, hs []func(string)) {
		cbsaved.RunAll(cross(cur), hs)
	})
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
