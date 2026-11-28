// PKGPATH: gno.land/r/test/indirectcurlocal
package indirectcurlocal

// Indirect cur-call through a local variable to a same-package crossing
// function must succeed. Preprocess can't statically bind the variable, so
// the same-realm check runs at runtime. Unlike func32.gno, the variable
// holds a non-nil func value, so the call reaches that check and passes.
// Companion to zrealm_curcall_indirect_external.gno, which verifies the
// runtime rejects the cross-realm case.

func myHandler(cur realm) {
	println("called")
}

func main(cur realm) {
	p := myHandler
	p(cur)
	println("done")
}

// Output:
// called
// done
