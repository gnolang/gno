// PKGPATH: gno.land/r/zrealm_trust_direct_memo
// The same t72b method is reached first from t72a's code (trusted), then from
// the realm's own interface call (not trusted): the second runs read-only.
package zrealm_trust_direct_memo

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
