// PKGPATH: gno.land/r/v3al2host
package v3al2host

import (
	"gno.land/r/tests/vm/v3al2store"
)

// Phase B-min+ Layer-2 demonstration.
//
// v3al2store.Token is a *v3al2helper.Holder stored in
// /r/tests/vm/v3al2store. When /r/v3al2host accesses Token.Reflect()
// — a /p/-declared method — the substrate's Layer-2 storage-receiver
// borrow shifts m.Realm to the Token's storage realm
// (/r/tests/vm/v3al2store). The v3a walk reads m.Realm via the
// LastRealm chain, so:
//   - Self()  resolves to /r/tests/vm/v3al2store (where the code
//             runs, post Layer-2 shift)
//   - Caller() resolves to /r/v3al2host (this realm)
//
// B-min (Layer-1 only) would have missed this because Token.Reflect's
// fv.PkgPath is /p/nt/v3al2helper (not /r/), so a PkgPath-based
// detector would skip the frame. B-min+'s m.Realm-shift detector
// catches it via the LastRealm chain.

func main() {
	caller, self := v3al2store.Token.Reflect()
	println("caller.PkgPath:", caller.PkgPath())
	println("self.PkgPath:  ", self.PkgPath())
}

// Output:
// caller.PkgPath: gno.land/r/v3al2host
// self.PkgPath:   gno.land/r/tests/vm/v3al2store
