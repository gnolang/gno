// PKGPATH: gno.land/r/demo/crossextra
package crossextra

// cross with a second argument that is not chain.Coins is rejected at
// preprocess time; the Go shim accepts any extra args, so only the
// preprocessor reports it.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross(cur, cur)))
}

// Error:
// gno.land/r/demo/crossextra/zrealm_cross_extra.gno:13:17-32: the second argument of cross must be a chain.Coins value

