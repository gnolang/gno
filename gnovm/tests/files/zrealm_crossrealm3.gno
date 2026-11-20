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
// cannot allocate gno.land/r/tests/vm.TestRealmObject in realm gno.land/r/crossrealm

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_crossrealm3.gno:15:5: cannot hold a value of gno.land/r/tests/vm.TestRealmObject outside its realm (use a pointer)
