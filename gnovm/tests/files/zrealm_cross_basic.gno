// PKGPATH: gno.land/r/demo/crossbasic
package crossbasic

// Happy path for cross(rlm): the explicit form of bare `cross` is
// semantically identical when rlm IS the current cur. Both forms
// mint a fresh cur whose prev points to the caller's frame.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	// Bare cross: classic form.
	println("bare cross:   ", callee(cross(cur)))
	// Explicit cross(cur): identical end state.
	println("cross(cur):  ", callee(cross(cur)))
}

// Output:
// bare cross:    gno.land/r/demo/crossbasic
// cross(cur):   gno.land/r/demo/crossbasic
