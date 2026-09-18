// PKGPATH: gno.land/r/demo/curshadowrange
package curshadowrange

// zrealm_cur_legal.gno shows `for _, cur := range` is legal to DECLARE; this
// pins that the declared shadow still cannot be passed to a crossing helper.

func probe(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	for _, cur := range []realm{cur.Previous()} {
		probe(cur)
	}
}

// Error:
// gno.land/r/demo/curshadowrange/zrealm_cur_shadow_range.gno:13:3-13: only the `cur` argument of a containing crossing function maybe passed by cross-call
