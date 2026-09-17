// PKGPATH: gno.land/r/demo/curother
package curother

// Realm-typed bindings named `cur` that are NOT a crossing parameter keep
// compiling: a package-level var, a named result, and a block-scoped local are
// not any frame's identity, so the write rules leave them alone. Only the
// parameter is fixed.

var pkgCur realm

func setPkgCur(x int, r realm) {
	pkgCur = r
	pkgCur = nil
}

func namedResult(x int) (cur realm) {
	cur = nil
	return cur
}

func rangeLocal(x int, r realm) {
	cur := r
	for _, cur = range []realm{r} {
	}
	_ = cur
}

func main() {
	setPkgCur(0, nil)
	_ = namedResult(0)
	rangeLocal(0, nil)
	println("ok")
}

// Output:
// ok
