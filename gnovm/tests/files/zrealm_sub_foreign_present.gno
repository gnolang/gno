// PKGPATH: gno.land/r/demo/subpresent
package subpresent

import (
	"chain"

	"gno.land/r/tests/vm/subhost"
)

// Cross-realm sub-identity presentation: this realm mints a sub of its
// OWN cur and crosses into a DIFFERENT realm (subhost). The callee reads
// the presented identity via cur.Previous() and must see:
//   - PkgPath() = the synthesized "<this host>:dao/9"
//   - Address() = chain.DerivePkgSubAddr(<this host>, "dao/9")
//   - Subpath() = "dao/9" (the canonical "am I a sub, of what" accessor,
//     resolved THROUGH the cross — this is the end-to-end check that
//     Subpath() works on a sub reached via cur.Previous() in a foreign
//     callee, not only on a locally-held token)
//   - IsUserCall() = false (a sub-token cross is never a direct user
//     call — load-bearing so payment guards don't treat a programmatic
//     sub-identity as a user)
//
// and unsafe.{Previous,Current}Realm() must agree at both heights.
//
// zrealm_sub1.gno covers the same parity but crossing WITHIN one
// package; this file crosses a real package boundary, which is the
// scenario every downstream consumer (boards, banker, grc20) actually
// sees.
//
// Two-step mint-then-cross: cross(...) requires a bare realm identifier.
func main(cur realm) {
	// Self-check the callee's DerivePkgSubAddr recomputation target.
	println("expected sub addr:",
		chain.DerivePkgSubAddr("gno.land/r/demo/subpresent", "dao/9"))
	sub := cur.Sub("dao/9")
	subhost.Observe(cross(sub))
}

// Output:
// expected sub addr: g1kug6thlz8wsc7nrs82j3urf58nv6j676tjal7r
// observed prev.PkgPath: gno.land/r/demo/subpresent:dao/9
// observed prev.Subpath: dao/9
// observed prev.IsUserCall: false
// prev.Address == DerivePkgSubAddr(host, subpath): true
// unsafe prev parity: true
// unsafe cur parity: true
