// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// A foreign-realm-typed package var carries the type's declared
// PkgID (authority = /r/tests/vm) but lives in the declaring
// package's storage block (owner = /r/crossrealm). Writes from
// /r/tests/vm code via a cross-call are still rejected because the
// existing readonly check uses owner-realm, not authority-realm.
// Authority and storage diverge under type-driven stamping; the
// system stays conservative and blocks at the readonly layer.
var somevalue tests.TestRealmObject

func main(cur realm) {
	tests.ModifyTestRealmObject(cross(cur), &somevalue)
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): t<VPBlock(1,1)>.Field
