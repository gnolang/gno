// PKGPATH: gno.land/r/launderh
package launderh

// Attack H: ATTACKER-PROVIDED INTERFACE IMPL WITH PRIMITIVE
// UNDERLYING TYPE.
//
// The borrow rule's receiver-borrow logic in PushFrameCall checks
// `recv.GetFirstObject(m.Store) != nil`. For struct/array/map/etc.
// receivers, this is the first reachable Object, whose PkgID stamp
// is what gets borrowed-to. For a PRIMITIVE-UNDERLYING receiver
// (int, bool, string, etc.) the value lives in tv.N, not tv.V —
// GetFirstObject returns nil (ownership.go default branch).
//
// Pre-fix: the receiver-borrow logic skipped and m.Realm stayed at
// whatever the caller had set, including a borrowed-from-victim
// m.Realm. An attacker-supplied EvilMutator (int-underlying) Run
// dispatched from inside a victim-borrowed /p/-method body
// inherited victim authority and committed the write.
//
// Fix: when recv has no Object (primitive-underlying defined type),
// anchor m.Realm to the receiver type's DECLARING PACKAGE — its
// only stable identity. The allocator's stamping context is left
// at the caller's natural realm via setRealmAuthorityOnly so the
// method body's local allocations still stamp correctly. Under
// this fix EvilMutator.Run runs with m.Realm = /p/launderattack,
// and the write to i.Field with base.PkgID = laundervictim trips
// the standard readonly check.
//
// Setup:
//   - /p/launderpkg defines `Mutator interface { Run(*Immutable) }`
//     and `func (o *Object) UseMutator(target *Immutable, m Mutator)`
//     which dispatches `m.Run(target)`.
//   - /p/launderattack defines `EvilMutator int` with a Run method
//     writing target.Field.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	gp.UseMutator(imm, launderattack.EvilMutator(0))
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,1)>.Field
