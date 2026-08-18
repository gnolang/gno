// PKGPATH: gno.land/r/demo/crossdefercross
package crossdefercross

// `defer f(cross, ...)` is rejected at preprocess time. Same reason as
// cross — doOpReturnCallDefers doesn't mint a fresh cur.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	defer callee(cross(cur))
}

// Error:
// gno.land/r/demo/crossdefercross/zrealm_cross_defercross.gno:12:8-26: crossing call `cross(rlm)` cannot be deferred directly — the deferred-call dispatch path doesn't mint a fresh cur. Wrap in a closure: `defer func() { fn(cross(rlm), args...) }()`
