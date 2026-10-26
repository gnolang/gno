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
// The position-0 rule is the one that fires here, because `c` resolves to
// realm as well. That is the point: the rules now see through the alias at
// every position, not only the one that first exposed the gap.

type R = realm

func f(c R, cur R) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	f(cur, cur.Previous())
}

// Error:
// gno.land/r/demo/curalias/zrealm_cur_alias.gno:19:1-19: a crossing function's first realm argument must have name `cur`
