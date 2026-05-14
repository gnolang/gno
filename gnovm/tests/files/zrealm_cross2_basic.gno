// PKGPATH: gno.land/r/demo/cross2basic
package cross2basic

// Happy path for cross2(rlm): the explicit form of bare `cross` is
// semantically identical when rlm IS the current cur. Both forms
// mint a fresh cur whose prev points to the caller's frame.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	// Bare cross: classic form.
	println("bare cross:   ", callee(cross))
	// Explicit cross2(cur): identical end state.
	println("cross2(cur):  ", callee(cross2(cur)))
}

// Output:
// bare cross:    gno.land/r/demo/cross2basic
// cross2(cur):   gno.land/r/demo/cross2basic
