// PKGPATH: gno.land/r/zrealm_trust_direct_chain
// The realm imports only t72m; t72m imports t72a, which imports t72b. t72a's
// code makes the interface call, and t72a imports t72b, so the write is allowed.
package zrealm_trust_direct_chain

import "gno.land/p/demo/tests/t72m"

var buf = []byte("....")

func Put(cur realm) { t72m.Run(buf) }

func main(cur realm) {
	Put(cross(cur))
	println("buf:", string(buf))
}

// Output:
// buf: B...
