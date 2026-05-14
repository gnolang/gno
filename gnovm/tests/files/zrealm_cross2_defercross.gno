// PKGPATH: gno.land/r/demo/cross2defercross
package cross2defercross

// `defer f(cross, ...)` is rejected at preprocess time. Same reason as
// cross2 — doOpReturnCallDefers doesn't mint a fresh cur.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	defer callee(cross)
}

// Error:
// gno.land/r/demo/cross2defercross/zrealm_cross2_defercross.gno:12:8-21: crossing call (`cross` or `cross2(rlm)`) cannot be deferred directly — the deferred-call dispatch path doesn't mint a fresh cur. Wrap in a closure: `defer func() { fn(cross, args...) }()`
