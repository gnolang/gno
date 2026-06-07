// PKGPATH: gno.land/r/launderrh4
package launderrh4

// /r/-DECLARED victim — substitution attack via foreign-realm
// construction.
//
// If an attacker could construct a fresh *launderrvictim.Immutable
// in /r/launderrh4, they could pass it to victim APIs that store
// caller-supplied values, then later mutate the impostor at will
// (the impostor would be stamped /r/launderrh4 = attacker's own
// realm).
//
// Blocked by the construction-time check
// (alloc.checkConstructionTime): allocating a /r/-declared type
// outside its declaring realm panics at the composite literal
// site, before the impostor ever exists.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	fake := &launderrvictim.Immutable{Field: "fake"}
	_ = fake
}

// Error:
// cannot allocate gno.land/r/tests/vm/launderrvictim.Immutable in realm gno.land/r/launderrh4
