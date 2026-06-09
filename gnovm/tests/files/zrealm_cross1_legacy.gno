// PKGPATH: gno.land/r/demo/cross1legacy
package cross1legacy

// cross1 is a legacy sentinel preserved as a migration aid for code
// originally written against the bare-`cross` form. It lowers to the
// same with-cross AST shape as the compiler-synthesized `.origin`
// (Args[0]=nil, WithCross=true) so the runtime mints the new cur via
// callingCurOrOrigin → buildOriginRealm. Migrate `cross1` → `cross(rlm)`
// once the in-scope realm value is identified.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println("cross1: ", callee(cross1))
	println("cross(cur):", callee(cross(cur)))
}

// Output:
// cross1:  gno.land/r/demo/cross1legacy
// cross(cur): gno.land/r/demo/cross1legacy
