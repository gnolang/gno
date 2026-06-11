// PKGPATH: gno.land/r/launderh
package launderh

// Attack H: ATTACKER-PROVIDED INTERFACE IMPL WITH PRIMITIVE
// UNDERLYING TYPE.
//
// PushFrameCall's receiver-storage borrow (borrow rule #2) checks
// `recv.GetFirstObject(m.Store) != nil`. For struct/array/map/etc.
// receivers it returns the first reachable Object whose PkgID stamp
// is what gets borrowed-to. For a PRIMITIVE-UNDERLYING receiver
// (int, bool, string, etc.) the value lives in tv.N, not tv.V —
// GetFirstObject returns nil and borrow rule #2 silently doesn't fire.
//
// In the laundervictim setup (data declared in /p/launderpkg), the
// caller is already running with m.Realm borrowed to /r/laundervictim
// (the outer UseMutator call hit borrow rule #2 on the *Object receiver). When
// the inner m.Run(target) interface-dispatches to EvilMutator.Run,
// borrow rule #2 silently no-ops and Run inherits the victim-borrowed m.Realm.
// Write to i.Field commits under victim authority.
//
// This shape is exploitable ONLY when the victim's logic data is
// declared in /p/ (here, Immutable lives in /p/launderpkg) so an
// attacker /p/ realm can implement `Run(*Immutable)`. If Immutable
// were declared in /r/laundervictim, /p/launderattack could not
// import it and could not implement the interface — the attack would
// be structurally impossible. Recommended practice: realms declare
// their own logic data types in /r/, not in /p/.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	println("before:", laundervictim.ReadImm())

	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	gp.UseMutator(imm, launderattack.EvilMutator(0))

	println("after: ", laundervictim.ReadImm())
}

// Output:
// before: original-imm
// after:  pwnd-via-iface
