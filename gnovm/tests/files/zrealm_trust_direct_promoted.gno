// PKGPATH: gno.land/r/zrealm_trust_direct_promoted
// t72p.P embeds t72c.C; the realm imports t72p, not t72c. Put promoted from
// t72c counts as t72c's code, so through an interface it is refused.
package zrealm_trust_direct_promoted

import "gno.land/p/demo/tests/t72p"

var buf = []byte("....")

func Put(cur realm) {
	var p interface{ Put([]byte) } = t72p.P{}
	p.Put(buf)
}

func main(cur realm) {
	Put(cross(cur))
	println("buf:", string(buf))
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): b<VPBlock(1,1)>[(const (0 int))]
