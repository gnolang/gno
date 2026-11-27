// PKGPATH: gno.land/r/zrealm_fvshare_deep
// R2a: sharing a very long pointer chain must not overflow the Go stack.
package zrealm_fvshare_deep

import "gno.land/r/tests/vm/fvshare"

type N struct{ next *N }

func main(cur realm) {
	var head *N
	for i := 0; i < 3000000; i++ {
		head = &N{next: head}
	}
	println("built")
	fvshare.Take(cross(cur), head)
	println("ok")
}

// Output:
// built
// ok
