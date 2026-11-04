// PKGPATH: gno.land/r/demo/curshadowif
package curshadowif

// Same as zrealm_cur_shadow_call.gno through an if-init: the shadow is owned
// by the IfStmt, so passing it to a crossing helper is refused.

func probe(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	if cur := cur.Previous(); cur != nil {
		probe(cur)
	}
}

// Error:
// gno.land/r/demo/curshadowif/zrealm_cur_shadow_if.gno:13:3-13: only the `cur` argument of a containing crossing function maybe passed by cross-call
