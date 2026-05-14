// PKGPATH: gno.land/r/demo/cross2deferwrap
package cross2deferwrap

// The closure-wrapper workaround for deferred crossing calls. The
// outer defer is on a non-crossing closure call; inside the closure
// body, the cross-call goes through the normal precall path and
// mints the fresh cur correctly.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	// defer the closure; inside, do the cross-call.
	defer func() {
		println("from defer:", callee(cross))
	}()
	println("body done")
}

// Output:
// body done
// from defer: gno.land/r/demo/cross2deferwrap
