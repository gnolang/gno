// PKGPATH: gno.land/r/launderrh_conv_slice_su
package launderrh_conv_slice_su

// CONVERSION VECTOR 8: Convert []launderpkg.Bare (foreign-readonly
// slice) to []LocalBare where LocalBare has the same underlying
// struct shape. If Gno permits this slice-to-slice conversion (Go
// 1.20+ allows it for slices of types with identical underlying
// types when wrapped in matching named types), the attacker could
// then iterate and write through LocalBare's method or field.
//
// Two probes:
//   A. []Bare -> []LocalBare (named target — Case 2 triggers since
//      LocalBare is /r/-declared but Case 1 fires first on src).
//   B. []Bare -> []struct{Field string} (anonymous target — Case 1
//      should still fire on source side).
//
// Hypothesis: both blocked at Case 1 of doOpConvert.

import (
	"gno.land/p/demo/tests/launderpkg"
	"gno.land/r/tests/vm/launderrvictim"
)

// LocalBare has identical underlying type to launderpkg.Bare.
type LocalBare struct {
	Field string
}

func (l *LocalBare) Set(s string) { l.Field = s }

func main(cur realm) {
	s := launderrvictim.GetBareSlice()
	// Probe A: convert to named /r/-declared slice element type.
	// In Go this requires either same underlying or specific rules.
	// In gno, both element types are defined types with same
	// underlying — Go disallows direct conversion []Bare->[]LocalBare
	// without an intermediate step; gno may or may not. The Case 1
	// check should fire either way before any element write.
	conv := []LocalBare(s)
	_ = conv
	conv[0].Set("pwnd-via-slice-conv")
	println("after:", launderrvictim.ReadBareSlice0())
	// silence unused-import
	_ = launderpkg.Bare{}
}

// TypeCheckError:
// gno.land/r/launderrh_conv_slice_su/zrealm_launder_rdata_conv_slice_same_underlying.gno:39:22: cannot convert s (variable of type []launderpkg.Bare) to type []LocalBare

// Error:
// gno.land/r/launderrh_conv_slice_su/zrealm_launder_rdata_conv_slice_same_underlying.gno:39:10-24: cannot convert s<VPBlock(1,1)> (of type []gno.land/p/demo/tests/launderpkg.Bare) to type []gno.land/r/launderrh_conv_slice_su.LocalBare
