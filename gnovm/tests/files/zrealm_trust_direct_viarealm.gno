// PKGPATH: gno.land/r/zrealm_trust_direct_viarealm
// t72b is imported only by another realm, t72q: its method runs read-only.
package zrealm_trust_direct_viarealm

import "gno.land/r/tests/vm/t72q"

var buf = []byte("....")

func Put(cur realm) {
	p := t72q.NewB()
	p.Put(buf)
}

func main(cur realm) {
	Put(cross(cur))
	println("buf:", string(buf))
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): b<VPBlock(1,1)>[(const (0 int))]
