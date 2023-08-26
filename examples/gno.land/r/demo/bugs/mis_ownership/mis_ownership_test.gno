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
}
