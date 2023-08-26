package mis_ownership

import "gno.land/r/demo/bugs/steal_ownership"

var (
	x   = uint32(42)
	ptr = &x
)

func init() {
	steal_ownership.Steal(ptr)
}
