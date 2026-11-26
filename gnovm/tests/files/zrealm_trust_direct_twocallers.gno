// PKGPATH: gno.land/r/zrealm_trust_direct_twocallers
// The same t72b method, called through an interface first by t72a's code
// (which imports t72b) and then by the realm's code (which does not), writes
// the first time and is refused the second.
package zrealm_trust_direct_twocallers

import "gno.land/p/demo/tests/t72a"

var buf = []byte("....")

func Put(cur realm) {
	p := t72a.NewB()
	t72a.Apply(p, buf)
	println("trusted:", string(buf))
	buf[0] = '.'
	p.Put(buf)
}

func main(cur realm) {
	Put(cross(cur))
	println("buf:", string(buf))
}

// Output:
// trusted: B...

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,1)>[(const (0 int))]
