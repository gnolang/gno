// PKGPATH: gno.land/r/launderrh_conv_iface_explicit
package launderrh_conv_iface_explicit

// CONVERSION VECTOR 2b: EXPLICIT cast to interface{} via type-conversion
// expression. Unlike the implicit assignment in conv_iface_box.gno,
// this form goes through doOpConvert, which applies Case 1 to refuse
// foreign-readonly source values regardless of the target type
// (including interface targets).
//
// Hypothesis: panics at the conversion site.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	w := launderrvictim.GetWithBarePtr()
	// Explicit cast (type-conversion expression) to interface{}.
	box := interface{}(w.Inner)
	_ = box
}

// Error:
// illegal conversion of readonly or externally stored value
