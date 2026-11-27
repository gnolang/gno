// PKGPATH: gno.land/r/zrealm_trust_direct_control
// As _deep, but the realm imports t72b directly: its method writes.
package zrealm_trust_direct_control

import (
	"gno.land/p/demo/tests/t72a"
	"gno.land/p/demo/tests/t72b"
)

var buf = []byte("....")

func Put(cur realm) {
	var p t72a.Putter = t72b.W{}
	p.Put(buf)
}

func main(cur realm) {
	Put(cross(cur))
	println("buf:", string(buf))
}

// Output:
// buf: B...
