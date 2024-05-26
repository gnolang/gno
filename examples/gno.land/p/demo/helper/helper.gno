package helper

import (
	"std"
)

// PrevRealmIsOwner checks if the caller of the function is the Realm's owner
func PrevRealmIsOwner(owner std.Address) bool {
	return std.GetOrigCaller() == owner
}
