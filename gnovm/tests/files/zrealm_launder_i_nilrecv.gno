// PKGPATH: gno.land/r/launderi
package launderi

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Attack I: NIL-POINTER-RECEIVER variant of Attack H.
//
// Where Attack H exploits primitive-underlying receivers
// (GetFirstObject returns nil via the ownership.go default branch),
// Attack I exploits POINTER-RECEIVER methods called on a nil pointer
// (PointerValue with Base==nil → GetBase returns nil →
// GetFirstObject returns nil). Same borrow rule #2 anchor gap, different
// shape: EvilNilRecv has a pointer-receiver method whose body does
// not dereference the receiver, so calling
// `(*EvilNilRecv)(nil).Run(imm)` from inside a borrowed-to-victim
// body inherits victim authority and the write commits.
//
// Same preconditions as Attack H: exploitable only when victim
// logic data is /p/-declared. Documented for completeness.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	println("before:", laundervictim.ReadImm())

	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	// Pass a NIL *EvilNilRecv as the Mutator. Run's body doesn't
	// deref the recv, so this doesn't immediately nil-panic — it
	// just writes target.Field.
	var nilEvil *launderattack.EvilNilRecv
	gp.UseMutator(imm, nilEvil)

	println("after: ", laundervictim.ReadImm())
}

// Output:
// before: original-imm

// Error:
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,1)>.Field
