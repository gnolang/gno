// PKGPATH: gno.land/r/launderk
package launderk

// Attack K: same shape as H but the interface method takes a
// `[]*orig.Immutable` parameter (slice of foreign-/p/ pointers)
// instead of a bare `*orig.Immutable`.
//
// The Attack H/I fix's `hasForeignPPtrParam` predicate only checks
// whether any parameter is itself a `*foreign-/p/-type`. A slice
// (or map, channel, struct, etc.) of those is NOT directly a
// pointer parameter, so the predicate skips. m.Realm inherits the
// borrowed-from-victim authority, and the slice element write
// inside the impl body commits.
//
// Tests whether the predicate needs to walk composite parameter
// types to find embedded foreign-/p/-pointers.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	gp.UseSliceMutator(imm, launderattack.EvilSliceMutator(0))
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): s<VPBlock(1,1)>[(const (0 int))].Field
