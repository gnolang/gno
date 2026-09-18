// PKGPATH: gno.land/r/stdlibgw
package stdlibgw

// Regression for the DidUpdate stdlib transient no-op on the NameExpr
// (package-global) write path. A (test-)stdlib top-level function writing
// its OWN package global via a bare-identifier assignment (counter++) must
// succeed when called from a /r/ realm — stdlib runs with the caller's
// authority and may mutate its own transient (never-persisted) state.
//
// The write goes: isExternalRealm (HIV branch — package globals are heap
// items — returns "not external") → DidUpdate, where po (stdlib-stamped) !=
// m.Realm (/r/). Without the DidUpdate stdlib no-op this hits the
// "invariant violation" backstop. (import4 covers the SelectorExpr field-
// write path, math/rand's p.lo; this covers the NameExpr global path.)

import "filetests/extern/stdlibglobalwrite"

func main(cur realm) {
	println(stdlibglobalwrite.Bump())
	println(stdlibglobalwrite.Bump())
}

// Output:
// 1
// 2
