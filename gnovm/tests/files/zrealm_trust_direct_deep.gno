// PKGPATH: gno.land/r/zrealm_trust_direct_deep
// The realm imports t72a, which imports t72b. A t72b method reached through
// the realm's own interface call is one hop too far: it runs read-only.
package zrealm_trust_direct_deep

import "gno.land/p/demo/tests/t72a"

var buf = []byte("....")

func Put(cur realm) {
	p := t72a.NewB()
	p.Put(buf)
}

func main(cur realm) {
	Put(cross(cur))
	println("buf:", string(buf))
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,1)>[(const (0 int))]
