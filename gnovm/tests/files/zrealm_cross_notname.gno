// PKGPATH: gno.land/r/demo/crossnotname
package crossnotname

// cross requires a bare realm-typed NameExpr — not an arbitrary
// realm-typed expression. Restricting to a name lets the runtime
// resolve via Block.GetPointerTo without re-entering the eval loop,
// and matches the realistic usage `cross(cur)` / `cross(rlm)`.

func realmHelper(cur realm) realm {
	return cur
}

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	// cross of a function-call expression — rejected by preprocessor.
	println(callee(cross(realmHelper(cur))))
}

// Error:
// gno.land/r/demo/crossnotname/zrealm_cross_notname.gno:19:17-40: cross argument must be a bare realm-typed identifier (a name, not an expression)
