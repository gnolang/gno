package mis_ownership

import "gno.land/r/demo/bugs/steal_ownership"

var (
	x     = uint32(42)
	slice = []uint32{4242}
	ptr   = &x
)

func init() {
	// steal_ownership.Steal(ptr) // at some point, this was allowed and prevented mutations of the object, maybe it's still present but it's another corner case

	steal_ownership.StealSlice(slice)
}
