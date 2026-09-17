// PKGPATH: gno.land/r/demo/curalias
package curalias

// The `cur` naming rules must look at the RESOLVED parameter type, not at how
// the type was spelled. They used to test `p.Type.(*NameExpr).Name == "realm"`
// while IsCrossing() compared the resolved type against gRealmType — so an
// alias slipped between them: the function was still crossing, but neither
// naming rule ran. That bought a crossing function a second, arbitrary,
// caller-supplied realm parameter literally named `cur`, which could then be
// forwarded by bare name. No assignment and no address-of anywhere, so none of
// the write-shape rules could ever see it.

// The position-0 rule is the one that fires here, because `c` resolves to
// realm as well — which is the point: the rules see through the alias at
// every position now, not just the one the attack used.

type R = realm

func f(c R, cur R) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	f(cur, cur.Previous())
}

// Error:
// gno.land/r/demo/curalias/zrealm_cur_alias.gno:19:1-19: a crossing function's first realm argument must have name `cur`
