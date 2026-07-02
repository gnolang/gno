// PKGPATH: gno.land/r/demo/subrealm
package subrealm

// Sub-of-sub is blocked by two independent guards. This test observes
// the first to fire: the receiver's host pkgpath is the synthesized
// "host:subpath", so the colon-shape check rejects it before the HIV
// guard runs. The strict-HIV guard (receiver's own HIV must be the
// topmost crossing Cur, which a sub-token's never is — even while its
// relaxed IsCurrent() reports true, printed here) is the load-bearing
// one; it's exercised directly by zrealm_sub_stale.gno. If Sub() ever
// routed through the IsCurrent relaxation instead, that test breaks.
func main(cur realm) {
	sub := cur.Sub("a")
	println("IsCurrent:", sub.IsCurrent())
	defer func() {
		println("recovered:", recover())
	}()
	sub.Sub("b")
}

// Output:
// IsCurrent: true
// recovered: Sub: receiver pkgpath is already synthesized or invalid
