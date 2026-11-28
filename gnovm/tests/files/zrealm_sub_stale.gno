// PKGPATH: gno.land/r/demo/subrealm
package subrealm

// Strict-guard staleness (recoverable): inside a crossed callee, the
// caller's cur is no longer the topmost crossing frame's Cur, so Sub()
// on it panics — recoverably, since Sub uses m.PanicString. Also shows
// a freshly-minted sub-token crossing successfully while its parent is
// live.
func spend(cur realm) {}

func callee(cur realm, outer realm) {
	defer func() {
		println("recovered:", recover())
	}()
	outer.Sub("x")
}

func main(cur realm) {
	callee(cross(cur), cur)
	sub := cur.Sub("y")
	spend(cross(sub))
	println("live cross ok")
}

// Output:
// recovered: Sub: receiver is not the live cur (stale, sibling, or a sub-token)
// live cross ok
