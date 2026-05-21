// PKGPATH: gno.land/r/caller
// MAXALLOC: 100000000
package caller

import "gno.land/r/tests/vm/crossrealm_d"

// A foreign realm cannot mutate caller-allocated []byte even when
// the slice is unreal (just allocated, not yet persisted). The byte
// slice's underlying ArrayValue is stamped with caller's PkgID at
// allocation; Layer 1 borrow on the /r/-declared MutateBytes
// activates m.Realm = crossrealm_d; the write site (bz[0] = 0xff)
// fires the readonly taint because the underlying PkgID != m.Realm.ID.
//
// Pre-interrealm-v2 this was open: the unreal slice had no PkgID, so
// the readonly check short-circuited and the foreign mutation went
// through. Allocation-time PkgID stamping closes the gap.

func main(cur realm) {
	bz := []byte{1, 2, 3}
	crossrealm_d.MutateBytes(bz) // panics inside MutateBytes
	println(bz[0])
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): bz<VPBlock(1,0)>[(const (0 int))]
