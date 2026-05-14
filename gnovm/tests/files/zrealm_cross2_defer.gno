// PKGPATH: gno.land/r/demo/cross2defer
package cross2defer

// `defer f(cross2(rlm), ...)` is rejected at preprocess time.
// doOpReturnCallDefers doesn't go through doOpPrecall, so the
// deferred callee never gets a freshly minted cur — it would crash
// at runtime. Wrap in a closure to defer cleanly.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	defer callee(cross2(cur))
}

// Error:
// gno.land/r/demo/cross2defer/zrealm_cross2_defer.gno:14:8-27: crossing call (`cross` or `cross2(rlm)`) cannot be deferred directly — the deferred-call dispatch path doesn't mint a fresh cur. Wrap in a closure: `defer func() { fn(cross, args...) }()`
