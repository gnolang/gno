// PKGPATH: gno.land/r/zrealm_trust_direct_memo_rev
// The reverse of _memo: the realm's own interface call (not trusted) comes
// first and is recovered, then t72a's (trusted) writes, then the realm's
// call is refused again.
package zrealm_trust_direct_memo_rev

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
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,1)>[(const (0 int))]
