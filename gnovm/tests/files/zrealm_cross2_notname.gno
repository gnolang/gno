// PKGPATH: gno.land/r/demo/cross2notname
package cross2notname

// cross2 requires a bare realm-typed NameExpr — not an arbitrary
// realm-typed expression. Restricting to a name lets the runtime
// resolve via Block.GetPointerTo without re-entering the eval loop,
// and matches the realistic usage `cross2(cur)` / `cross2(rlm)`.

func realmHelper(cur realm) realm {
	return cur
}

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	// cross2 of a function-call expression — rejected by preprocessor.
	println(callee(cross2(realmHelper(cur))))
}

// Error:
// gno.land/r/demo/cross2notname/zrealm_cross2_notname.gno:19:17-41: cross2 argument must be a bare realm-typed identifier (a name, not an expression)
