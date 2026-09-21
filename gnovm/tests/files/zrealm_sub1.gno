// PKGPATH: gno.land/r/demo/subrealm
package subrealm

import "chain/runtime/unsafe"

// Crossing with a sub-token: the callee sees the sub-identity via
// cur.Previous(), and unsafe.{Current,Previous}Realm agree at every
// comparable position (v2 §5.4 parity). The callee — entered via a
// sub-token cross — can mint subs of its own fresh cur (legit nesting
// across frames; only sub-of-sub on the same token is blocked).
//
// Note the two-step mint-then-cross: cross(...) requires a bare
// realm-typed identifier, so cross(cur.Sub(x)) composed inline is
// rejected at preprocess.
func callee(cur realm) {
	println("prev.PkgPath:", cur.Previous().PkgPath())
	println("prev.Address:", cur.Previous().Address())
	println("unsafe parity prev:",
		string(cur.Previous().Address()) == string(unsafe.PreviousRealm().Address()) &&
			cur.Previous().PkgPath() == unsafe.PreviousRealm().PkgPath())
	println("unsafe parity cur:",
		string(cur.Address()) == string(unsafe.CurrentRealm().Address()) &&
			cur.PkgPath() == unsafe.CurrentRealm().PkgPath())
	inner := cur.Sub("inner")
	println("nested mint:", inner.PkgPath())
}

func main(cur realm) {
	sub := cur.Sub("dao/1")
	callee(cross(sub))
}

// Output:
// prev.PkgPath: gno.land/r/demo/subrealm#dao/1
// prev.Address: g1jy9ajp3mqdrcsklfgeafqpfr0nn2ypejnexzv6
// unsafe parity prev: true
// unsafe parity cur: true
// nested mint: gno.land/r/demo/subrealm#inner
