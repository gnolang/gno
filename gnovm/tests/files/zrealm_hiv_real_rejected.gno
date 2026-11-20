// PKGPATH: gno.land/r/hivrealrejected
package hivrealrejected

// persistedcap's init stores a closure built by /p/ p_closurecap.MakeCounter;
// its captured `c` is a real heap item owned by persistedcap. Called from
// this realm, the closure runs view-only, so its write to `c` is refused.

import "gno.land/r/tests/vm/persistedcap"

func main(cur realm) {
	fn := persistedcap.GetCounter()
	println(fn())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): c<~VPBlock(1,1)>
