// PKGPATH: gno.land/r/demo/crossifacestale
package crossifacestale

// Runtime IsCurrent-strict catches a stale rlm passed through
// INTERFACE DISPATCH. This is the realistic shape of the cur-leak
// the broader migration exists to prevent: a foreign-realm impl
// would receive cur via an interface method, but cross ensures the
// rlm being handed over IS the current crossing frame's Cur.
//
// Construction: realm-typed value threaded through a cross-call
// chain (main→A→B→A). At the second A invocation, the original
// main's cur is no longer topmost. Invoking `iface.Cross(cross(outerCur))`
// from second-A panics on IsCurrent-strict.

type Crosser interface {
	Cross(cur realm) string
}

type impl struct{}

func (impl) Cross(cur realm) string {
	return cur.PkgPath()
}

func A(cur realm, outerCur realm, c Crosser, depth int) {
	if depth > 0 {
		// outerCur is main's cur, no longer topmost.
		// cross through interface dispatch — runtime rejects.
		println(c.Cross(cross(outerCur)))
		return
	}
	B(cross(cur), outerCur, c)
}

func B(cur realm, outerCur realm, c Crosser) {
	A(cross(cur), outerCur, c, 1)
}

func main(cur realm) {
	var c Crosser = impl{}
	A(cross(cur), cur, c, 0)
}

// Error:
// cross: rlm is not the current cur (stale capture or sibling frame)
