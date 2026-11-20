// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// Foreign minting is rejected: crossrealm_b receives this realm's live
// cur but its code, entered without crossing, runs read-only and not as
// this realm, so the caller-pkgpath guard refuses to mint a sub-identity
// in a namespace the executing code doesn't own.
func main(cur realm) {
	crossrealm_b.TrySubOn(0, cur)
}

// Error:
// Sub: caller is not operating in the receiver's namespace
