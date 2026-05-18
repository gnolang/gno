// PKGPATH: gno.land/r/launderi
package launderi

// Attack I: NIL-POINTER-RECEIVER variant of Attack H.
//
// Where Attack H exploited primitive-underlying receivers
// (GetFirstObject returns nil via the ownership.go default branch),
// Attack I exploits POINTER-RECEIVER methods called on a nil pointer
// (PointerValue with Base==nil → GetBase returns nil →
// GetFirstObject returns nil). Same anchor gap, different shape.
//
// EvilNilRecv has a pointer-receiver method whose body does not
// dereference the receiver. Calling `(*EvilNilRecv)(nil).Run(imm)`
// from inside a borrowed-to-victim body would inherit victim
// authority — UNLESS PushFrameCall handles the nil-pointer-base case.
//
// Note: the fix from Attack H gates on `recv.T.(*DeclaredType)`. For
// a pointer receiver, recv.T is *PointerType (not DeclaredType), so
// the Attack H fix's branch does NOT fire. This test confirms
// whether the nil-pointer variant is also closed or whether a
// separate guard is needed.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	// Pass a NIL *EvilNilRecv as the Mutator. Run's body doesn't
	// deref the recv, so this doesn't immediately nil-panic — it
	// just writes target.Field.
	var nilEvil *launderattack.EvilNilRecv
	gp.UseMutator(imm, nilEvil)
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,1)>.Field
