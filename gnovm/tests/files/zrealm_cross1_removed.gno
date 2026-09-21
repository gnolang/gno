// PKGPATH: gno.land/r/demo/cross1removed
package cross1removed

// cross1 was a temporary sentinel for the bare-`cross` migration era.
// It has been removed; this test asserts the name no longer resolves.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross1))
}

// Error:
// gno.land/r/demo/cross1removed/zrealm_cross1_removed.gno:12:17-23: name cross1 not declared

// TypeCheckError:
// gno.land/r/demo/cross1removed/zrealm_cross1_removed.gno:12:17: undefined: cross1
