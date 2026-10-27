package counter

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestIncrementDecrement(cur realm, t *testing.T) {
	start := GetCount()

	Increment(cross(cur))
	Increment(cross(cur))
	uassert.Equal(t, start+2, GetCount())

	Decrement(cross(cur))
	uassert.Equal(t, start+1, GetCount())
}
