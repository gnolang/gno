// PKGPATH: gno.land/r/demo/curshadowcall
package curshadowcall

// A shadowing `cur` may hold another realm's value, but it must never reach the
// one call shape that promotes a value into a frame identity: the cur-call
// check resolves the name to its declaring block, and a `{ }` block is not a
// crossing function.

func probe(cur realm) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	{
		cur := cur.Previous()
		probe(cur)
	}
}

// Error:
// gno.land/r/demo/curshadowcall/zrealm_cur_shadow_call.gno:16:3-13: only the `cur` argument of a containing crossing function maybe passed by cross-call
