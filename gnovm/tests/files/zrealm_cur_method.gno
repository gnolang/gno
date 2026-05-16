// PKGPATH: gno.land/r/curmethod
package curmethod

// Bound crossing methods: cur lives at block[1] (block[0] is the receiver),
// so the doOpCall non-cross entry fix uses curIdx=1. Verifies that:
//   - non-cross call `t.M(cur)` inherits caller's cur (prev unchanged)
//   - cross-call `t.M(cross)` opens a new chain level (prev = caller's cur)
type T struct{ name string }

func (t T) M(cur realm) realm {
	return cur
}

func main(cur realm) {
	t := T{name: "rex"}
	mCurNonCross := t.M(cur)
	println("non-cross: t.M(cur) returns same realm as caller cur:", mCurNonCross == cur)

	mCurCross := t.M(cross2(cur))
	println("cross: t.M(cross) returns a different realm (new capture):", mCurCross != cur)
	println("cross: returned cur same pkg path:", mCurCross.PkgPath() == cur.PkgPath())
	// After cross-call, the inner cur's prev is the caller's cur (main's).
	println("cross: returned.Previous() == main's cur:", mCurCross.Previous() == cur)
}

// Output:
// non-cross: t.M(cur) returns same realm as caller cur: true
// cross: t.M(cross) returns a different realm (new capture): true
// cross: returned cur same pkg path: true
// cross: returned.Previous() == main's cur: true
