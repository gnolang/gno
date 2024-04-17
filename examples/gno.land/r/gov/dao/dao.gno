package govdao

import "std"

func IsMember(addr std.Address) bool {
	// XXX: implement
	return true
}

func AssertIsMember(addr std.Address) {
	if !IsMember(addr) {
		panic("caller is not member of govdao")
	}
}

// XXX: add other helpers
