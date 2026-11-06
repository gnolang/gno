// PKGPATH: gno.land/r/test/indirectcurlocal
package indirectcurlocal

// Indirect cur-call through a local variable to a same-package crossing
// function must succeed. Unlike func32.gno, the variable holds a non-nil
// func value, so this exercises the bind-succeeds path. Companion to
// zrealm_curcall_indirect_external.gno, which verifies the runtime rejects
// the cross-realm case.

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
