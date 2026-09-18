// PKGPATH: gno.land/r/demo/curother
package curother

// The declaration rule keys on the pair (name `cur`, resolved type `realm`).
// Realm-typed bindings under any other name — a package-level var, a named
// result, a local, a range variable — are ordinary realm values and stay
// legal. A `cur` of any other type is refused: the name is reserved.

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

func main() {
	setPkgRealm(0, nil)
	_ = namedResult(0)
	rangeLocal(0, nil)
	println("ok")
}

// Output:
// ok
