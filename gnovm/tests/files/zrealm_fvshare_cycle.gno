// PKGPATH: gno.land/r/zrealm_fvshare_cycle
// R2a: sharing a value that points back at itself ends.
package zrealm_fvshare_cycle

import "gno.land/r/tests/vm/fvshare"

var (
	x any
	a *any
	b any
)

func main(cur realm) {
	x = &x
	fvshare.Take(cross(cur), x)
	b = &a
	var bb any = b
	a = &bb
	fvshare.Take(cross(cur), a)
	println("ok")
}

// Output:
// ok
