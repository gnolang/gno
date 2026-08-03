// PKGPATH: gno.land/r/launderrh
package launderrh

// Attack L against /r/-DECLARED victim data.
//
// Victim launderrvictim declares Immutable in /r/. UseAnyMutator
// boxes gImm (*launderrvictim.Immutable) as any and dispatches
// /p/launderpkg.AnyMutator. /p/launderattack.EvilAnyMutator
// implements AnyMutator with `Run(x any)` that type-asserts to
// *launderpkg.Immutable.
//
// Expected: the type assertion fails — the boxed value is
// *launderrvictim.Immutable, not *launderpkg.Immutable. The
// attacker's /p/-declared impl cannot name *launderrvictim.Immutable
// in any type-assert position (can't import /r/), so it has no
// path to a writable view of the field. Panic at the assertion.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	launderrvictim.UseAnyMutator(launderattack.EvilAnyMutator(0))

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// *gno.land/r/tests/vm/launderrvictim.Immutable is not of type *gno.land/p/demo/tests/launderpkg.Immutable
