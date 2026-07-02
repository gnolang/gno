// PKGPATH: gno.land/r/demo/subrealm
package subrealm

// Sub-of-sub is structurally blocked: the strict entry guard demands
// the receiver's own HIV be the topmost crossing frame's Cur, which a
// sub-token's never is — even while its relaxed IsCurrent() reports
// true. This divergence is load-bearing; if Sub() ever routes through
// the IsCurrent relaxation, this test breaks.
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
