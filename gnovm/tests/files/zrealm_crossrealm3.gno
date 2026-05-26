// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// Method receivers don't change the result for foreign-realm-typed
// vars: under the new rule, somevalue's StructValue is /r/tests/vm-
// stamped, and borrow rule #2 shifts m.Realm to /r/tests/vm inside
// Modify(). But the receiver pointer's Base is /r/crossrealm's
// package block (which holds the slot), and IsReadonlyBy checks
// that Base's PkgID, not the slot value's. /r/crossrealm != /r/tests/vm
// → rejected.
var somevalue tests.TestRealmObject

func main(cur realm) {
	somevalue.Modify()
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): t<VPBlock(1,0)>.Field
