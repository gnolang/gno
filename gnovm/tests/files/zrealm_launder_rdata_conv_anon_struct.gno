// PKGPATH: gno.land/r/launderrh_conv_anon_struct
package launderrh_conv_anon_struct

// CONVERSION VECTOR 3: Convert *launderpkg.Bare (foreign-readonly,
// declared /p/-type) to *struct{Field string} — an ANONYMOUS struct
// pointer that's "neither" /p/ nor /r/. The Case 2 target check only
// fires for *DeclaredType targets; an anonymous struct pointer's
// element type is a raw *StructType, NOT a DeclaredType. So Case 2
// doesn't apply.
//
// Hypothesis: Case 1 still applies (source is foreign-readonly), so
// the conversion is blocked regardless of target shape.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	w := launderrvictim.GetWithBarePtr()
	anon := (*struct{ Field string })(w.Inner)
	anon.Field = "pwnd-via-anon-target"
	println("after:", launderrvictim.ReadBarePtr())
}

// Error:
// illegal conversion of readonly or externally stored value
