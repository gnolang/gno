// PKGPATH: gno.land/r/launderl
package launderl

// Attack L: same shape as H but the interface method takes `any`
// (empty interface) instead of `*orig.Immutable`. The caller (UseAnyMutator)
// boxes target into the any. The attacker's impl type-asserts back
// to *Immutable and writes the field.
//
// The Attack H/I fix's static predicate (reachesForeignPPtr) walks
// pointer/slice/array/map/chan/struct composites but does NOT walk
// into interface types — they have no compile-time element type.
// At the call site, the runtime value boxed into `any` is a foreign
// /p/-pointer, but the static signature doesn't reveal that.
//
// Tests whether the predicate needs to fire on interface-typed
// parameters too (over-broad — every interface param would be
// considered "potentially writable foreign") or whether some
// runtime check is needed.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

// *** STILL OPEN ***
// Treating interface-typed params as foreign-pointer-bearing is
// over-broad (breaks zrealm7 / avl + types.String.Less). Awaiting a
// narrower defense — likely runtime arg-walk only when receiver
// lacks an Object anchor.

func main(cur realm) {
	println("before:", laundervictim.ReadImm())

	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	gp.UseAnyMutator(imm, launderattack.EvilAnyMutator(0))

	println("after: ", laundervictim.ReadImm())
}

// Output:
// before: original-imm
// after:  pwnd-via-any-arg
