// PKGPATH: gno.land/r/zrealm_trust_direct_dispatch
// t72a's code makes the interface call and t72a imports t72b: it writes.
package zrealm_trust_direct_dispatch

import "gno.land/p/demo/tests/t72a"

var buf = []byte("....")

func Put(cur realm) {
	t72a.Apply(t72a.NewB(), buf)
}

func main(cur realm) {
	Put(cross(cur))
	println("buf:", string(buf))
}

// Output:
// buf: B...
