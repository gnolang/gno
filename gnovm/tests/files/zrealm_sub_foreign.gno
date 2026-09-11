// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// Foreign minting is rejected: crossrealm_b receives this realm's live
// cur (the strict guard passes — it IS the topmost crossing cur) but
// runs under borrow rule #1 with m.Realm = crossrealm_b, so the
// caller-pkgpath guard refuses to mint a sub-identity in a namespace
// the executing realm doesn't own. The panic surfaces from the
// borrowed foreign frame as an abort.
func main(cur realm) {
	crossrealm_b.TrySubOn(0, cur)
}

// Error:
// Sub: caller is not operating in the receiver's namespace
