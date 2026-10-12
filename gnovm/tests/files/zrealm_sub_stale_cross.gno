// PKGPATH: gno.land/r/demo/subrealm
package subrealm

// A sub-token goes stale once its parent frame is no longer the topmost
// crossing cur: crossing with it then fails. cross's guard is a raw
// (machine-fatal) panic — deliberately NOT recoverable — so this aborts
// the program rather than being caught.
func spend(cur realm) {}

func stale(cur realm, sub realm) {
	spend(cross(sub))
}

func main(cur realm) {
	sub := cur.Sub("y")
	stale(cross(cur), sub)
}

// Error:
// cross: rlm is not the current cur (stale capture or sibling frame)
