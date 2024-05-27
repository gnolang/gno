package pausable

import (
	"std"
)

func prevRealmIsOwner(addr std.Address) bool {
	return std.PrevRealm().Addr() == addr
}
