// PKGPATH: gno.land/r/demo/curdeclrange
package curdeclrange

// A range DEFINE declares a fresh binding in the range clause's own block.
// Over a []realm that binding is realm-typed, so the declaration is refused;
// the assign form is refused separately, as a write to the crossing
// parameter (zrealm_cur_range.gno).

func main(cur realm) {
	for _, cur := range []realm{cur.Previous()} {
		_ = cur
	}
	println("ok")
}

// Error:
// gno.land/r/demo/curdeclrange/zrealm_cur_decl_range.gno:10:2-12:3: a realm-typed `cur` may only be declared as the first parameter of a crossing function
