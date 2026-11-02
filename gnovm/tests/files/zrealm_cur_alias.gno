// PKGPATH: gno.land/r/demo/curalias
package curalias

// The `cur` naming rules must decide on the RESOLVED parameter type, not on how
// the type was spelled. They used to test `p.Type.(*NameExpr).Name == "realm"`
// while IsCrossing() compared the resolved type against gRealmType, and an
// alias fits between the two: the function is still crossing, but neither
// naming rule runs. That let a crossing function declare a SECOND realm
// parameter literally named `cur` — an ordinary caller-supplied argument that
// the cur-call provenance check would then accept by name, with no assignment
// and no address taken anywhere for the write-shape rules to see.
//
// `cur` is now reserved by name (StaticBlock.Reserve), so a second parameter
// named `cur` is refused before any type resolves; the alias rule is pinned
// here through position 0, where `c` resolves to realm and is not `cur`.

type R = realm

func f(c R, x R) {
	_ = x.IsCurrent()
}

func main(cur realm) {
	f(cur, cur.Previous())
}

// Error:
// gno.land/r/demo/curalias/zrealm_cur_alias.gno:19:1-17: a crossing function's first realm argument must have name `cur`
