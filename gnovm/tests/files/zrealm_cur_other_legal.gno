// PKGPATH: gno.land/r/demo/curother
package curother

// The declaration rule keys on the pair (name `cur`, resolved type `realm`).
// Realm-typed bindings under any other name — a package-level var, a named
// result, a local, a range variable — are ordinary realm values and stay
// legal, and a `cur` of a non-realm type stays legal too.

var pkgRealm realm

func setPkgRealm(x int, r realm) {
	pkgRealm = r
	pkgRealm = nil
}

func namedResult(x int) (rlm realm) {
	rlm = nil
	return rlm
}

func rangeLocal(x int, r realm) {
	rlm := r
	for _, rlm = range []realm{r} {
	}
	_ = rlm
}

func intCur(x int) {
	cur := x
	cur = 7
	p := &cur
	*p = 8
	for _, cur = range []int{9} {
	}
	println("int cur:", cur)
}

func main() {
	setPkgRealm(0, nil)
	_ = namedResult(0)
	rangeLocal(0, nil)
	intCur(0)
	println("ok")
}

// Output:
// int cur: 9
// ok
