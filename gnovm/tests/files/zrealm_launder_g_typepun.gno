// PKGPATH: gno.land/r/launderg
package launderg

// Attack G: TYPE-PUNNING LAUNDER.
//
// /p/launderpkg defines Immutable: a struct with no mutator method.
// /r/laundervictim stores a *Immutable, intending callers to be
// unable to write its Field.
//
// /p/launderattack defines Tamper: a different /p/ type with the SAME
// underlying struct layout, and adds a Set method. Without the Case 1
// source-side readonly check in doOpConvert, an attacker can convert
// *Immutable → *Tamper (the conversion is a pure type-tag change;
// the pointer still aliases victim's persisted memory) and then
// invoke Tamper.Set — PushFrameCall's borrow rule 2 routes m.Realm
// to laundervictim because the underlying *StructValue is still
// victim-stamped, and the write to o.Field inside Set runs with
// victim authority.
//
// The data's PkgID stays correctly stamped throughout; the launder
// is of the METHOD SET, not the stamp. The defense is Case 1 in
// doOpConvert: refuse conversion of a foreign-readonly source value.
// This blocks the conversion before any method dispatch can happen.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	imm := laundervictim.GetImm() // *launderpkg.Immutable
	tamper := launderattack.Convert(imm)
	tamper.Set("pwnd-via-typepun")
}

// Error:
// illegal conversion of readonly or externally stored value
