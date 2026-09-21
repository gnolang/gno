// Package subhost is a test fixture for realm.Sub sub-identity
// presentation across a real package boundary. It plays two roles:
//
//   - Host: ActAsSub mints a sub of its OWN live cur and crosses into an
//     observer with it.
//   - Observer: Observe reads the presented identity (cur.Previous) that
//     a caller crossed in with, and asserts unsafe.PreviousRealm parity.
//
// Consumers:
//   - gnovm/tests/files/zrealm_sub_foreign_present.gno crosses into
//     Observe directly with a sub minted in the caller's own realm, to
//     check cross-realm identity presentation + Subpath() through a cross.
//   - gno.land/pkg/integration/testdata/subrealm_run_parity.txtar enters
//     via MsgRun, crosses into ActAsSub, which mints a sub and crosses
//     into Observe — the MsgRun-entry counterpart to the MsgCall-entry
//     parity filetest zrealm_sub1.gno.
package subhost

import (
	"chain"
	"chain/runtime/unsafe"
)

const pkgPath = "gno.land/r/tests/vm/subhost"

// Observe reports what the immediate caller presented as its identity.
// The three cur.Previous() views (address, synthesized pkgpath, subpath)
// are the callee-side surface a sub-token must present; the two parity
// booleans assert unsafe.PreviousRealm() reads the SAME identity (v2
// §5.4: unsafe.* must agree with cur.Previous()). prevParity is the
// load-bearing assertion for the identity-chain walk; curParity guards
// height 0.
func Observe(cur realm) {
	prev := cur.Previous()
	println("observed prev.PkgPath:", prev.PkgPath())
	println("observed prev.Subpath:", prev.Subpath())
	println("observed prev.IsUserCall:", prev.IsUserCall())
	// Short-circuit the primary case FIRST: DerivePkgSubAddr panics on an
	// empty subpath, so it must not be evaluated when prev is a primary.
	println("prev.Address == DerivePkgSubAddr(host, subpath):",
		prev.Subpath() == "" ||
			string(prev.Address()) == string(chain.DerivePkgSubAddr(pkgPathOf(prev.PkgPath()), prev.Subpath())))
	println("unsafe prev parity:",
		string(prev.Address()) == string(unsafe.PreviousRealm().Address()) &&
			prev.PkgPath() == unsafe.PreviousRealm().PkgPath())
	println("unsafe cur parity:",
		string(cur.Address()) == string(unsafe.CurrentRealm().Address()) &&
			cur.PkgPath() == unsafe.CurrentRealm().PkgPath())
}

// pkgPathOf returns the host portion of a possibly-synthesized pkgpath,
// so DerivePkgSubAddr (which takes the host) can be recomputed from the
// callee's observation alone.
func pkgPathOf(p string) string {
	host, _, _ := chain.SplitPkgSubPath(p)
	return host
}

// ActAsSub mints a sub-identity of subhost's OWN live cur and crosses
// with it into Observe. Used by the MsgRun-entry parity test: the
// outermost entry is a `/e/<addr>/run` ephemeral, but subhost's cur is a
// first-class primary here, so cur.Sub is legal and the sub presents to
// Observe exactly as in the MsgCall path.
func ActAsSub(cur realm, subpath string) {
	sub := cur.Sub(subpath)
	Observe(cross(sub))
}
