package mis_ownership

import (
	"testing"
)

func TestMisOwnership(t *testing.T) {
	// by value
	x = 21

	// deref
	*ptr = 21

	// pointer change
	nv := uint32(21)
	ptr = &nv

	// slice access
	slice[0] = 2121 // this triggers `cannot modify external-realm or non-realm object`

	// slice change
	slice = []uint32{2121} // this works
}
