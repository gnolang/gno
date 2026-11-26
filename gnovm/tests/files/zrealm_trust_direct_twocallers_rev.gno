// PKGPATH: gno.land/r/zrealm_trust_direct_twocallers_rev
// As _twocallers in the other order: the realm's call is refused (and
// recovered), t72a's call then writes, and the realm's call is refused again.
package zrealm_trust_direct_twocallers_rev

import "gno.land/p/demo/tests/t72a"

var buf = []byte("....")

func try(p t72a.Putter) {
	defer func() { println("recovered:", recover() != nil) }()
	p.Put(buf)
}

func Put(cur realm) {
	p := t72a.NewB()
	try(p)
	println("untrusted:", string(buf))
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
// recovered: true
// untrusted: ....
// trusted: B...

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): b<VPBlock(1,1)>[(const (0 int))]
