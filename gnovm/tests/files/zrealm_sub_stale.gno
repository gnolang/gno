// PKGPATH: gno.land/r/demo/subrealm
package subrealm

// Strict-guard staleness: inside a crossed callee, the caller's cur is
// no longer the topmost crossing frame's Cur, so Sub() on it panics —
// same for cross() of a sub-token minted before the frame shifted.
func callee(cur realm, outer realm) {
	defer func() {
		println("recovered:", recover())
	}()
	outer.Sub("x")
}

func spend(cur realm) {}

func main(cur realm) {
	callee(cross(cur), cur)
	// A sub-token minted here is cross-usable here...
	sub := cur.Sub("y")
	spend(cross(sub))
	println("live cross ok")
	// ...but goes stale once its parent frame is no longer topmost.
	stale(cross(cur), sub)
}

func stale(cur realm, sub realm) {
	defer func() {
		println("recovered:", recover())
	}()
	spend(cross(sub))
}

// Error:
// cross: rlm is not the current cur (stale capture or sibling frame)
