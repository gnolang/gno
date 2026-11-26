// PKGPATH: gno.land/r/zrealm_saved_escaped_container
// R2a attack: R.Do builds a local table and a caller's crossing callback
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
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
