// PKGPATH: gno.land/r/demo/subrealm
package subrealm

// Sub-tokens cannot be persisted (refusePersistRealmHIV; the origin
// carve-out excludes sub-tokens by subpath).
var stash realm

func main(cur realm) {
	stash = cur.Sub("vault")
}

// Error:
// cannot persist realm value: realm values are ephemeral and tied to a call frame
