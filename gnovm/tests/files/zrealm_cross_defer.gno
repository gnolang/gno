// PKGPATH: gno.land/r/demo/crossdefer
package crossdefer

// `defer f(cross(rlm), ...)` is rejected at preprocess time.
// doOpReturnCallDefers doesn't go through doOpPrecall, so the
// deferred callee never gets a freshly minted cur — it would crash
// at runtime. Wrap in a closure to defer cleanly.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	defer callee(cross(cur))
}

// Error:
// gno.land/r/demo/crossdefer/zrealm_cross_defer.gno:14:8-26: crossing call `cross(rlm)` cannot be deferred directly — the deferred-call dispatch path doesn't mint a fresh cur. Wrap in a closure: `defer func() { fn(cross(rlm), args...) }()`
